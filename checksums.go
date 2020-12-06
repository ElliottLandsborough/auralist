package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strconv"

	"github.com/kalafut/imohash"
	"github.com/vcaesar/murmur"
	"golang.org/x/crypto/ssh"
)

// Get crc32 hash as an integer
func hashFileCrc32(filePath string) (int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return -1, err
	}
	defer file.Close()
	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, file); err != nil {
		return -1, err
	}
	hashInBytes := hash.Sum(nil)[:]
	CRC32String := hex.EncodeToString(hashInBytes)
	CRC32Int, err := strconv.ParseInt(CRC32String, 16, 64)

	if _, err := io.Copy(hash, file); err != nil {
		return -1, err
	}

	return CRC32Int, nil
}

// Get imo hash of file
func hashFileImo(filePath string) (string, error) {
	hash, err := imohash.SumFile(filePath)

	if err != nil {
		return "", err
	}

	hashStr := fmt.Sprintf("%x", hash)

	return hashStr, nil
}

// hashFileMd5 https://mrwaggel.be/post/generate-md5-hash-of-a-file-in-golang/
func hashFileMd5(filePath string) (string, error) {
	// Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	// Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	// Tell the program to call the following function when the current function returns
	defer file.Close()

	// Open a new hash interface to write to
	hash := md5.New()

	// Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	// Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	// Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil
}

// Generate murmur hash of string
func stringToMurmur(path string) uint32 {
	return murmur.Murmur3([]byte(path))
}

// HashStringMd5 generates an md5 hash from a string
func HashStringMd5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// hashFileMd5Remote gets the md5 hash of a file on the other end of an ssh connection
func hashFileMD5Remote(path string, sshClient *ssh.Client) string {
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

// hashFileSHA1Remote gets the sha1 hash of a file on the other end of an ssh connection
func hashFileSHA1Remote(path string, sshClient *ssh.Client) string {
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
