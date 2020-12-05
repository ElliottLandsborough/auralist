package main

import (
	"fmt"

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

// Create file row
func createFileRow(db *gorm.DB, file File) error {
	// Only insert when PathHash doesn't exist, otherwise update
	db.Create(&file)

	return nil
}

// Create file rows
func createFileRows(db *gorm.DB, files []File) error {
	// Only insert when PathHash doesn't exist, otherwise update
	db.Create(&files)

	return nil
}

func deleteAllFiles(db *gorm.DB) {
	db.Where("true").Delete(&File{})
}

func deleteAllTags(db *gorm.DB) {
	db.Where("true").Delete(&Tag{})
}
