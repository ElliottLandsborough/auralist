package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"gorm.io/gorm"
)

// PathInfo contains a full path and other basic info
type PathInfo struct {
	Path string // /home/ubuntu/Music/donk.mp3
	Size int64  // file size in bytes (maximum 4294967295, 4gb!)
}

// Iterate through files in directory
func indexFiles(db *gorm.DB) error {
	paths := make([]string, 0)

	fmt.Println("Collecting paths...")

	e := filepath.Walk(conf.SearchDirectory, func(path string, f os.FileInfo, err error) error {
		// Don't process directories
		if !f.IsDir() {
			paths = append(paths, path)
		}

		return err
	})

	if e != nil {
		panic(e)
	}

	fmt.Printf("Found %d files.\n", len(paths))

	handlePaths(paths, db)

	return nil
}

func handlePaths(paths []string, db *gorm.DB) error {
	// Queue size = double the cpu count
	queueLength := runtime.NumCPU() * 2
	// Limit queue to a file size also
	var fileSizeLimit int64 = 500000000
	// Initialize combined size
	var combinedSize int64 = 0
	// Init empty queue items slice
	queueItems := make([]PathInfo, 0)

	// Loop through all the paths
	for _, path := range paths {
		// Get filesize for this item
		size, err := getFileSizeInBytes(path)

		if err != nil {
			panic(err)
		}

		// Add size to combined size
		combinedSize = combinedSize + size

		path := PathInfo{
			Path: path,
			Size: size}

		// Add item to queue
		queueItems = append(queueItems, path)

		// If we hit the size limit, process
		if combinedSize > fileSizeLimit {
			createFileRows(db, processFilesAsync(queueItems))
			queueItems = nil
		}

		// If we hit the queue length limit, process
		if len(queueItems) >= queueLength {
			createFileRows(db, processFilesAsync(queueItems))
			queueItems = nil
		}
	}

	// If we got here and we still have items left, process
	if len(queueItems) > 0 {
		createFileRows(db, processFilesAsync(queueItems))
	}

	return nil
}

func processFilesAsync(paths []PathInfo) []File {
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

			pi := paths[i]

			file, err := createFile(pi)

			if err != nil {
				panic(err)
			}

			fmt.Printf("Found %s files.\n", pi.Path)

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

func createFile(pi PathInfo) (File, error) {
	Crc32, err := hashFileCrc32(pi.Path)

	if err != nil {
		return File{}, err
	}

	HostName, err := os.Hostname()

	if err != nil {
		panic(err)
	}

	file := File{
		PathHash:           stringToMurmur(pi.Path),
		FileName:           filepath.Base(pi.Path),
		Path:               strings.ReplaceAll(pi.Path, conf.SearchDirectory, ""),
		Base:               conf.SearchDirectory,
		FileSizeBytes:      pi.Size,
		ExtensionLowerCase: trimLeftChars(strings.ToLower(filepath.Ext(pi.Path)), 1),
		Crc32:              Crc32,
		HostName:           HostName}

	return file, nil

}
