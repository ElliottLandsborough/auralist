package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
		FullPathHash:       stringToMurmur(path),
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

func deleteAllTags(db *gorm.DB) {
	db.Where("true").Delete(&Tag{})
}
