package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Initialization routine.
func init() {
	// Retrieve config options.
	conf = getConf()
}

// File info
type File struct {
	ID                 uint
	FullPath           string
	FileName           string
	ExtensionLowerCase string `gorm:"index"`
	Crc32              uint   `gorm:"index"`
	Crc32WithoutTags   uint   `gorm:"index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func listFiles(*gorm.DB) ([]string, error) {
	fileList := make([]string, 0)
	e := filepath.Walk(conf.SearchDirectory, func(path string, f os.FileInfo, err error) error {
		fmt.Println(path)
		return err
	})

	if e != nil {
		panic(e)
	}

	return fileList, nil
}

func getDB() (*gorm.DB, error) {
	dsn := "root:password@tcp(127.0.0.1:3306)/auralist?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	return db, err
}
