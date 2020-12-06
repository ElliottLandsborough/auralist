package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	scp "github.com/bramvdbogaerde/go-scp"
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

func getSSHClient(server string, user string) *ssh.Client {
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

func copyFiles() {
	server := conf.SSHServer + ":" + conf.SSHPort
	user := conf.SSHUser
	remotePath := conf.RemotePath

	sshClient := getSSHClient(server, user)

	scpClient, err := scp.NewClientBySSH(sshClient)
	if err != nil {
		fmt.Println("Error creating new SSH session from existing connection", err)
	}

	// Open a file
	f, _ := os.Open("/proc/cpuinfo")

	// Close client connection after the file has been copied
	defer scpClient.Close()

	// Close the file after it has been copied
	defer f.Close()

	remoteFilePath := remotePath + "cpuinfo"

	// Usage: CopyFile(fileReader, remotePath, permission)
	err = scpClient.CopyFile(f, remoteFilePath, "0644")

	// Get md5 of remote file
	md5sum := md5RemoteFile(remoteFilePath, sshClient)

	// Get sha1 of remote file
	sha1sum := sha1RemoteFile(remoteFilePath, sshClient)

	fmt.Println(md5sum)
	fmt.Println(sha1sum)

	if err != nil {
		fmt.Println("Error while copying file ", err)
	}
}

func md5RemoteFile(path string, sshClient *ssh.Client) string {
	session := getSSHSession(sshClient)
	defer session.Close()

	command := "/usr/bin/md5sum -z \"" + path + "\""

	output, err := remoteRun(command, session)

	if len(output) == 0 {
		panic("MD5 failed.")
	}

	if err != nil {
		panic(err)
	}

	// get first 32 chars
	md5sum := output[0:31]

	return md5sum
}

// 1937c5c26c25fba7e07e48eb8ec39dcf04e033a5
func sha1RemoteFile(path string, sshClient *ssh.Client) string {
	session := getSSHSession(sshClient)
	defer session.Close()

	command := "/usr/bin/sha1sum -z \"" + path + "\""

	fmt.Println(command)

	output, err := remoteRun(command, session)

	if len(output) == 0 {
		panic("SHA1 failed.")
	}

	if err != nil {
		panic(err)
	}

	// get first 32 chars
	sha1sum := output[0:39]

	return sha1sum
}

// e.g. output, err := remoteRun("whoami", session)
func remoteRun(command string, session *ssh.Session) (string, error) {
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	err := session.Run(command)

	stdOut := stdoutBuf.String()

	return stdOut, err
}
