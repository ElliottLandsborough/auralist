package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

// Iterate through files in directory
func listFiles(db *gorm.DB) ([]string, error) {
	fileList := make([]string, 0)
	e := filepath.Walk(conf.SearchDirectory, func(path string, f os.FileInfo, err error) error {
		// Don't process directories
		if !f.IsDir() {
			allowedExtentions := []string{".mp3", ".flac", ".ogg"}

			// is the extension allowed?
			if stringInSlice(filepath.Ext(path), allowedExtentions) {
				rowError := createFileRow(db, path)

				fmt.Println(path)

				if rowError != nil {
					panic(rowError)
				}
			}
		}

		return err
	})

	if e != nil {
		panic(e)
	}

	return fileList, nil
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
