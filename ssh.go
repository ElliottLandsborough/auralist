package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/alessio/shellescape.v1"
	"gorm.io/gorm"
)

var (
	sshClient *ssh.Client
)

// Initialization routine.
func init() {
	// Retrieve config options.
	conf = getConf()

	// Seed random numbers with nanotime
	rand.Seed(time.Now().UnixNano())
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

func waitForSSHClient() {
	for {
		log.Println("Connecting...")
		// Connect to server
		client, err := getSSHClient()

		if err != nil {
			log.Println("SSH Dial failed")
			log.Println("Sleeping for 10s")
			time.Sleep(10 * time.Second)
			continue
		}

		log.Println("Connected!")

		sshClient = client

		break
	}

}

func getSSHClient() (*ssh.Client, error) {
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
		log.Println("Failed to dial: " + err.Error())
		return nil, err
	}

	return client, nil
}

// Will keep trying forever
func getSSHSession() *ssh.Session {
	// try to get new session
	session, err := sshClient.NewSession()

	if err != nil {
		log.Println("Failed to create session: " + err.Error())

		waitForSSHClient()
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
func fileExistsOnRemoteServer(path string) bool {
	session := getSSHSession()
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

func fileMatchOnRemoteServer(localFullPath string, remoteFullPath string, file File, db *gorm.DB) (bool, error) {
	// Check remote location for file, if it exists already
	if fileExistsOnRemoteServer(remoteFullPath) {
		// Get an md5 hash of it
		localMD5, err := hashFileMd5(localFullPath)

		if err != nil {
			return false, err
		}

		remoteMD5, err := hashFileMD5Remote(remoteFullPath)

		if err != nil {
			return false, err
		}

		// If local md5 matches remote md5
		if localMD5 == remoteMD5 {
			file.Md5 = remoteMD5
			file.VerifiedAt = time.Now()
			db.Save(&file)
			return true, nil
		}
	}

	return false, nil
}

// recursively create directories required
func createDirectoryRecursiveRemote(path string) bool {
	session := getSSHSession()
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
func directoryExistsRemote(path string) bool {
	session := getSSHSession()
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
func createEmptyFileRemote(path string) bool {
	session := getSSHSession()
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

func createZeroFileOnRemoteServerIfNotExists(remoteFullPath string) bool {
	// Check remote location for file, if it does not exist
	if !fileExistsOnRemoteServer(remoteFullPath) {
		// Check if directory already exists
		if directoryExistsRemote(filepath.Dir(remoteFullPath)) {
			if createEmptyFileRemote(remoteFullPath) {
				return true
			}
		}
		// Try to create directories
		if createDirectoryRecursiveRemote(filepath.Dir(remoteFullPath)) {
			if createEmptyFileRemote(remoteFullPath) {
				return true
			}
		}
	}

	return false
}

func copyFileRemote(source string, destination string) bool {
	session := getSSHSession()
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

func copyFromOldFolderIfExists(file File, localFullPath string, remoteFullPath string, db *gorm.DB) (bool, error) {
	// Get remote hostname
	remoteHostName, err := remoteRun("hostname", getSSHSession())

	if err != nil {
		log.Println("Could not get remote hostname.")
		return false, err
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

		fm, err := fileMatchOnRemoteServer(localFullPath, remoteOldFullPath, file, db)

		if err != nil {
			log.Println("Error Getting match between local and remote")
			return false, err
		}

		// Does local md5 match remote old path md5?
		if fm {
			// Create directories on remote server
			createDirectoryRecursiveRemote(remoteFullPath)
			// Copy file one remote from old location to new location
			if copyFileRemote(remoteOldFullPath, remoteFullPath) {
				// Copy success
				return true, nil
			}
		}
	}

	return false, nil
}

func uploadFile(localFullPath string, remoteFullPath string, file File, db *gorm.DB) (bool, error) {
	// time.Duration is in nanoseconds, int64. 1 hour = 1 * 60 * 60 * 1000 * 1000 * 1000
	var timeOut time.Duration = 10 * 60 * 1000 * 1000 * 1000 // 10 mins

	scpClient, err := scp.NewClientBySSHWithTimeout(sshClient, timeOut)
	if err != nil {
		log.Println("Error creating new SSH session from existing connection", err)
	}

	// Open a file
	f, _ := os.Open(localFullPath)

	if !createDirectoryRecursiveRemote(filepath.Dir(remoteFullPath)) {
		log.Println("Could not create remote directory " + filepath.Dir(remoteFullPath))

		return false, nil
	}

	log.Println("Uploading `" + filepath.Base(remoteFullPath) + "`")

	// Close client connection after the file has been copied
	defer scpClient.Close()

	// Close the file after it has been copied
	defer f.Close()

	// Define chunk size in bytes
	var chunkSize int64 = 100 * 1000 * 1000 // 100 mb

	// File is larger than chunksize
	if file.FileSizeBytes > chunkSize {
		upload, err := uploadFileInChunks(localFullPath, remoteFullPath, file, chunkSize)
		if upload {
			return true, nil
		}

		log.Println("Error while uploading chunked file ", localFullPath)
		return false, err
	}

	// Usage: CopyFile(fileReader, remotePath, permission)
	err = scpClient.Copy(f, shellescape.Quote(remoteFullPath), "0644", file.FileSizeBytes)

	if err != nil {
		fmt.Println("Error while uploading whole file ", localFullPath)
		return false, err
	}

	return fileMatchOnRemoteServer(localFullPath, remoteFullPath, file, db)
}

func uploadFileInChunks(localFullPath string, remoteFullPath string, file File, chunkSize int64) (bool, error) {
	random64 := randSeq(64)

	pathPrefix := "/tmp/auralist.tmp." + random64 + ".part"

	f, err := os.Open(localFullPath)

	if err != nil {
		log.Println("Error opening local file during chunked upload: " + localFullPath)
		return false, err
	}

	defer f.Close()

	b := make([]byte, chunkSize)
	chunks := make([]string, 0)

	var chunkCount = 0

	for {
		bytesReadCount, err := f.Read(b)

		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}

			break
		}

		// generate filenames that end in '000000001', '000000002', ...
		chunk := b[:bytesReadCount]
		path := pathPrefix + fmt.Sprintf("%09d", chunkCount)
		chunks = append(chunks, path)

		err = writeChunkToRemoteTmpFile(chunk, bytesReadCount, path)

		if err != nil {
			log.Println("Error writing remote chunk")
			return false, err
		}

		chunkCount++
	}

	err = joinRemoteChunks(pathPrefix, remoteFullPath)

	if err != nil {
		log.Println("Error joining remote chunks")
		return false, err
	}

	err = deleteRemoteChunks(chunks)

	if err != nil {
		log.Println("Error deleting remote chunks: " + err.Error())
	}

	return true, nil
}

func joinRemoteChunks(pathPrefix string, remoteFullPath string) error {
	session := getSSHSession()
	defer session.Close()

	command := "cat " + pathPrefix + "* > " + shellescape.Quote(remoteFullPath)

	_, err := remoteRun(command, session)

	if err != nil {
		return err
	}

	return nil
}

func deleteRemoteChunks(chunks []string) error {
	for _, chunk := range chunks {
		session := getSSHSession()
		defer session.Close()

		command := "rm " + shellescape.Quote(chunk)

		_, err := remoteRun(command, session)

		if err != nil {
			return err
		}
	}

	return nil
}

func writeChunkToRemoteTmpFile(chunk []byte, chunkSize int, path string) error {
	session := getSSHSession()
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintln(w, "C0644", chunkSize, filepath.Base(path))
		fmt.Fprint(w, string(chunk))
		fmt.Fprint(w, "\x00")
	}()

	_, err := remoteRun("/usr/bin/scp -rt "+path, session)

	if err != nil {
		return err
	}

	return nil
}

func randSeq(n uint) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
