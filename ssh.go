package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

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

	stdOut := stdoutBuf.String()

	return stdOut, err
}
