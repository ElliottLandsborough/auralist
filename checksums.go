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
