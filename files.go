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
	fileQueueFlushLimit := 1
	fileQueue := make([]File, 0)
	e := filepath.Walk(conf.SearchDirectory, func(path string, f os.FileInfo, err error) error {
		// Don't process directories
		if !f.IsDir() {
			fmt.Println(path)

			file, err := createFile(path)

			if err != nil {
				panic(err)
			}

			fileQueue = append(fileQueue, file)

			if len(fileQueue) == fileQueueFlushLimit {
				flushFileQueue(fileQueue, db)
				fileQueue = make([]File, 0)
			}
		}

		return err
	})

	if e != nil {
		panic(e)
	}

	return fileList, nil
}

func flushFileQueue(files []File, db *gorm.DB) {
	for _, file := range files {
		//allowedExtentions := []string{".mp3", ".flac", ".ogg"}
		// is the extension allowed?
		//if stringInSlice(filepath.Ext(path), allowedExtentions) {

		rowError := createFileRow(db, file)

		if rowError != nil {
			panic(rowError)
		}
	}

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
