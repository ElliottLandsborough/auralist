package main

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/vcaesar/murmur"
)

// Gets db connection info
func getDSN() string {
	MysqlDatabase := conf.MysqlDatabase
	MysqlUser := conf.MysqlUser
	MysqlPass := conf.MysqlPass
	MysqlHost := conf.MysqlHost

	return fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8&parseTime=True&loc=Local",
		MysqlUser,
		MysqlPass,
		MysqlHost,
		MysqlDatabase)
}

// Gets a db connection
func getDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(getDSN()), &gorm.Config{})

	return db, err
}

// Create file row
func createFileRow(db *gorm.DB, path string) error {
	FileSizeBytes, err := getFileSizeInBytes(path)

	if err != nil {
		return err
	}

	Crc32, err := hashFileCrc32(path)

	if err != nil {
		return err
	}

	FileRow := File{
		FullPathHash:       generateFullPathHash(path),
		FullPath:           path,
		FileName:           filepath.Base(path),
		FileSizeBytes:      FileSizeBytes,
		ExtensionLowerCase: strings.ToLower(filepath.Ext(path)),
		Crc32:              Crc32}

	// Only insert when FullPathHash doesn't exist, otherwise update
	db.Create(&FileRow)

	return nil
}

func deleteAllFiles(db *gorm.DB) {
	db.Where("true").Delete(&File{})
}

// Gets a files size in bytes
func getFileSizeInBytes(path string) (int64, error) {
	fi, err := os.Stat(path)

	if err != nil {
		return -1, err
	}

	// get the size
	size := fi.Size()

	return size, nil
}

// Generate a hash of the path
func generateFullPathHash(path string) uint32 {
	return murmur.Murmur3([]byte(path))
}

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
