package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// Iterate through files in directory
func indexFiles(db *gorm.DB) error {
	paths := make([]string, 0)
	queueSize := 64

	e := filepath.Walk(conf.SearchDirectory, func(path string, f os.FileInfo, err error) error {
		// Don't process directories
		if !f.IsDir() {
			fmt.Println(path)

			paths, err = handlePaths(path, paths, queueSize, db)

			if err != nil {
				panic(err)
			}
		}

		return err
	})

	if e != nil {
		panic(e)
	}

	return nil
}

func handlePaths(path string, paths []string, queueSize int, db *gorm.DB) ([]string, error) {
	paths = append(paths, path)

	if len(paths) > queueSize {
		files := processFilesAsync(paths)
		createFileRows(db, files)
		paths = nil
	}

	return paths, nil
}

func processFilesAsync(paths []string) []File {
	pathsLength := len(paths)
	files := make([]File, 0)

	// Initialize wait group
	var wg sync.WaitGroup

	// How many items do we want to process concurrently
	wg.Add(pathsLength)

	// Each path...
	for i := 0; i < pathsLength; i++ {
		// Spawn thread
		go func(i int) {
			// When this thread finishes let the waitgroup know
			defer wg.Done()

			file, err := createFile(paths[i])

			if err != nil {
				panic(err)
			}

			files = append(files, file)
		}(i)
	}

	// Wait until processing has finished
	wg.Wait()

	return files
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

func createFile(path string) (File, error) {
	FileSizeBytes, err := getFileSizeInBytes(path)

	if err != nil {
		return File{}, err
	}

	Crc32, err := hashFileCrc32(path)

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
		HostName:           HostName}

	return file, nil

}
