package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/alessio/shellescape.v1"
	"gorm.io/gorm"
)

// Initialization routine.
func init() {
	// Retrieve config options.
	conf = getConf()
}

// Get private key into memory
func getPrivateKey() []byte {
	keyFile := conf.SSHKey

	file, err := os.Open(keyFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	b, err := ioutil.ReadAll(file)

	return b
}

// create human-readable SSH-key strings
func keyString(k ssh.PublicKey) string {
	return k.Type() + " " + base64.StdEncoding.EncodeToString(k.Marshal()) // e.g. "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY...."
}

func trustedHostKeyCallback(trustedKey string) ssh.HostKeyCallback {

	if trustedKey == "" {
		return func(_ string, _ net.Addr, k ssh.PublicKey) error {
			log.Printf("WARNING: SSH-key verification is *NOT* in effect: to fix, add this trustedKey: %q", keyString(k))
			return nil
		}
	}

	return func(_ string, _ net.Addr, k ssh.PublicKey) error {
		ks := keyString(k)
		if trustedKey != ks {
			return fmt.Errorf("SSH-key verification: expected %q but got %q", trustedKey, ks)
		}

		return nil
	}
}

func getSSHClient() *ssh.Client {
	server := conf.SSHServer + ":" + conf.SSHPort
	user := conf.SSHUser

	signer, _ := ssh.ParsePrivateKey(getPrivateKey())
	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: trustedHostKeyCallback(conf.SSHHostKey),
	}

	client, err := ssh.Dial("tcp", server, clientConfig)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	return client
}

func getSSHSession(client *ssh.Client) *ssh.Session {
	session, err := client.NewSession()

	if err != nil {
		panic("Failed to create session: " + err.Error())
	}

	return session
}

// e.g. output, err := remoteRun("whoami", session)
func remoteRun(command string, session *ssh.Session) (string, error) {
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	err := session.Run(command)

	stdOut := strings.TrimSpace(stdoutBuf.String())

	return stdOut, err
}

//
func fileExistsOnRemoteServer(path string, sshClient *ssh.Client) bool {
	session := getSSHSession(sshClient)
	defer session.Close()

	command := "test -f " + shellescape.Quote(path)

	_, err := remoteRun(command, session)

	// non zero output, file does not exist
	if err != nil {
		return false
	}

	// zero output, no errors, file exists
	return true
}

func fileMatchOnRemoteServer(localFullPath string, remoteFullPath string, sshClient *ssh.Client) bool {
	// Check remote location for file, if it exists already
	if fileExistsOnRemoteServer(remoteFullPath, sshClient) {
		// Get an md5 hash of it
		localMd5, err := hashFileMd5(localFullPath)

		if err != nil {
			panic(err)
		}

		remoteMd5 := hashFileMD5Remote(remoteFullPath, sshClient)

		// If local md5 matches remote md5
		if localMd5 == remoteMd5 {
			return true
		}
	}

	return false
}

// recursively create directories required
func createDirectoryRecursiveRemote(path string, sshClient *ssh.Client) bool {
	session := getSSHSession(sshClient)
	defer session.Close()

	command := "mkdir -p " + shellescape.Quote(path)

	_, err := remoteRun(command, session)

	// non zero output, file does not exist
	if err != nil {
		return false
	}

	// zero output, no errors, file exists
	return true
}

// Check if directory exists on remote server
func directoryExistsRemote(path string, sshClient *ssh.Client) bool {
	session := getSSHSession(sshClient)
	defer session.Close()

	command := "test -d " + shellescape.Quote(path)

	_, err := remoteRun(command, session)

	// non zero output, fail
	if err != nil {
		return false
	}

	// zero output, success
	return true
}

// Create zero-byte file
func createEmptyFileRemote(path string, sshClient *ssh.Client) bool {
	session := getSSHSession(sshClient)
	defer session.Close()

	command := "touch " + shellescape.Quote(path)

	_, err := remoteRun(command, session)

	// non zero output, file does not exist
	if err != nil {
		return false
	}

	// zero output, no errors, file exists
	return true
}

func createZeroFileOnRemoteServerIfNotExists(remoteFullPath string, sshClient *ssh.Client) bool {
	// Check remote location for file, if it does not exist
	if !fileExistsOnRemoteServer(remoteFullPath, sshClient) {
		// Check if directory already exists
		if directoryExistsRemote(filepath.Dir(remoteFullPath), sshClient) {
			if createEmptyFileRemote(remoteFullPath, sshClient) {
				return true
			}
		}
		// Try to create directories
		if createDirectoryRecursiveRemote(filepath.Dir(remoteFullPath), sshClient) {
			if createEmptyFileRemote(remoteFullPath, sshClient) {
				return true
			}
		}
	}

	return false
}

func copyFileRemote(source string, destination string, sshClient *ssh.Client) bool {
	session := getSSHSession(sshClient)
	defer session.Close()

	command := "cp " + shellescape.Quote(source) + " " + shellescape.Quote(destination)

	_, err := remoteRun(command, session)

	// non zero output, failed
	if err != nil {
		return false
	}

	// zero output, success
	return true
}

func copyFromOldFolderIfExists(file File, localFullPath string, remoteFullPath string, db *gorm.DB, sshClient *ssh.Client) bool {
	// Get remote hostname
	remoteHostName, err := remoteRun("hostname", getSSHSession(sshClient))

	if err != nil {
		panic("Could not get local hostname")
	}

	// Empty file
	potentialDuplicate := File{}

	// Check db for first result where crc32 match
	db.Where(&File{HostName: remoteHostName, Crc32: file.Crc32}).First(&potentialDuplicate)

	// If we have a CRC32 match on the remote server for this file (check in '')
	if len(potentialDuplicate.FileName) > 0 {
		log.Println("Matched old file `" + potentialDuplicate.FileName + "`")

		// generate path to file in old folder
		remoteOldFullPath := conf.RemoteOldPath + potentialDuplicate.Path

		// Does local md5 match remote old path md5?
		if fileMatchOnRemoteServer(localFullPath, remoteOldFullPath, sshClient) {
			// Create directories on remote server
			createDirectoryRecursiveRemote(remoteFullPath, sshClient)
			// Copy file one remote from old location to new location
			if !copyFileRemote(remoteOldFullPath, remoteFullPath, sshClient) {
				panic("Remote copy failed.")
			}
			// Copy success, move to next file
			return true
		}
	}
	return false
}

func uploadFile(localFullPath string, remoteFullPath string, file File, sshClient *ssh.Client) {
	// time.Duration is in nanoseconds, int64. 1 hour = 1 * 60 * 60 * 1000 * 1000 * 1000
	var timeOut time.Duration = 10 * 60 * 1000 * 1000 * 1000 // 10 mins

	scpClient, err := scp.NewClientBySSHWithTimeout(sshClient, timeOut)
	if err != nil {
		log.Println("Error creating new SSH session from existing connection", err)
	}

	// Open a file
	f, _ := os.Open(localFullPath)

	if !createDirectoryRecursiveRemote(filepath.Dir(remoteFullPath), sshClient) {
		panic("Could not create remote directory " + filepath.Dir(remoteFullPath))
	}

	log.Println("Uploading `" + filepath.Base(remoteFullPath) + "`")

	// Close client connection after the file has been copied
	defer scpClient.Close()

	// Close the file after it has been copied
	defer f.Close()

	// Define chunk size in bytes
	var chunkSize int64 = 1 * 1000 * 1000 // 100 mb

	// File is larger than chunksize
	if file.FileSizeBytes > chunkSize {
		uploadFileInChunks(localFullPath, remoteFullPath, file, chunkSize, sshClient)
	} else {
		// Usage: CopyFile(fileReader, remotePath, permission)
		err = scpClient.Copy(f, shellescape.Quote(remoteFullPath), "0644", file.FileSizeBytes)
	}

	if err != nil {
		fmt.Println("Error while copying file ", localFullPath)
	}

	if err != nil {
		panic(err)
	}
}

func uploadFileInChunks(localFullPath string, remoteFullPath string, file File, chunkSize int64, sshClient *ssh.Client) {
	md5, err := hashFileMd5(localFullPath)

	if err != nil {
		panic(err)
	}

	f, err := os.Open(localFullPath)

	if err != nil {
		panic("Error opening file" + localFullPath)
	}

	defer f.Close()

	b := make([]byte, chunkSize)
	chunkFiles := make([]string, 0)

	var chunkCount = 0

	for {
		bytesReadCount, err := f.Read(b)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}

			break
		}

		chunk := b[:bytesReadCount]
		path := "/tmp/auralist." + strconv.FormatInt(file.Crc32, 10) + ".part" + strconv.Itoa(chunkCount)
		chunkFiles = append(chunkFiles, path)

		fmt.Println(localFullPath)
		fmt.Println("MD5: " + md5)
		fmt.Println("Chunk: " + path)
		fmt.Println("Bytes: " + strconv.Itoa(bytesReadCount))
		writeChunkToRemoteTmpFile(chunk, bytesReadCount, path, sshClient)

		chunkCount++
	}

	joinRemoteChunkFiles(chunkFiles)
}

func joinRemoteChunkFiles(chunks []string) {
	for _, chunk := range chunks {
		fmt.Println(chunk)
	}
}

func writeChunkToRemoteTmpFile(chunk []byte, chunkSize int, path string, sshClient *ssh.Client) {
	session := getSSHSession(sshClient)
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		fmt.Fprintln(w, "C0644", chunkSize, shellescape.Quote(path))
		fmt.Fprint(w, string(chunk))
		fmt.Fprint(w, "\x00") // transfer end with \x00
	}()

	_, err := remoteRun("/usr/bin/scp -rt "+shellescape.Quote(path), getSSHSession(sshClient))

	if err != nil {
		fmt.Println(err)
	}
}

/*
func getSensibleFileSplitSize() int64 {
	// Get memory stats here
	memory, err := memory.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)

		// Could not get ram? assume 100
		return 100
	}

	// Find out how much a quarter of available ram is
	quarterOfRAMInBytes := memory.Free / 4

	// Round to nearest ten megabytes
	splitSizeInBytes := int64(math.Round(float64(quarterOfRAMInBytes)/1000/1000/10) * 1000 * 1000 * 10)

	// 1mb for now
	splitSizeInBytes = 1 * 1000 * 1000

	// Do we even have enough ram for this??
	if splitSizeInBytes < 50 {
		// Todo: this is hacky. Can probably manage memory better. Or use rsync...
		panic("Not enough ram.")
	}

	return 200
}
*/
