package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Initialization routine.
func init() {
	// Retrieve config options.
	conf = getConf()
}

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

func createFile(path string) (File, error) {
	FileSizeBytes, err := getFileSizeInBytes(path)

	if err != nil {
		return File{}, err
	}

	Crc32, err := hashFileCrc32(path)

	if err != nil {
		return File{}, err
	}

	ImoHash, err := hashFileImo(path)

	if err != nil {
		return File{}, err
	}

	HostName, err := os.Hostname()

	if err != nil {
		panic(err)
	}

	file := File{
		PathHash:           stringToMurmur(path),
		FileName:           filepath.Base(path),
		Path:               strings.ReplaceAll(path, conf.SearchDirectory, ""),
		Base:               conf.SearchDirectory,
		FileSizeBytes:      FileSizeBytes,
		ExtensionLowerCase: trimLeftChars(strings.ToLower(filepath.Ext(path)), 1),
		Crc32:              Crc32,
		ImoHash:            ImoHash,
		HostName:           HostName}

	return file, nil

}

// Create file row
func createFileRow(db *gorm.DB, file File) error {
	// Only insert when PathHash doesn't exist, otherwise update
	db.Create(&file)

	return nil
}

func deleteAllFiles(db *gorm.DB) {
	db.Where("true").Delete(&File{})
}

func deleteAllTags(db *gorm.DB) {
	db.Where("true").Delete(&Tag{})
}
