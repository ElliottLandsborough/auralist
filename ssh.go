package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
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

	command := "test -f \"" + path + "\""

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

	command := "mkdir -p\"" + path + "\""

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

	command := "test -d\"" + path + "\""

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

	command := "touch \"" + path + "\""

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

	command := "cp \"" + source + "\"" + "\"" + destination + "\""

	_, err := remoteRun(command, session)

	// non zero output, failed
	if err != nil {
		return false
	}

	// zero output, success
	return true
}
