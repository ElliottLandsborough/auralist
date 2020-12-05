package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// File with music in it
type File struct {
	ID                 uint
	FileName           string // donk.mp3
	Path               string // /home/ubuntu/Music/donk.mp3
	FullPathMd5        string `gorm:"index;size:32"`
	Base               string
	PathHash           uint32 `gorm:"index"` // murmur3(Path)
	FileSizeBytes      int64  // file size in bytes (maximum 4294967295, 4gb!)
	ExtensionLowerCase string `gorm:"index"`          // mp3
	Crc32              int64  `gorm:"index"`          // 321789321
	HostName           string `gorm:"index;size:256"` // max linux hostname size as per manpage
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Handles paths inputted into it. Rudimentary queue system
func handlePaths(paths []string, db *gorm.DB) error {
	// Queue size = double the cpu count
	queueLength := runtime.NumCPU() * 4
	// Limit queue to a file size also
	var fileSizeLimit int64 = 500000000
	// Initialize combined size
	var combinedSize int64 = 0
	// Init empty queue items slice
	queueItems := make([]PathInfo, 0)

	// Loop through all the paths
	for _, path := range paths {
		if fileIsInDatabase(path, db) == false {
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
	}

	// If we got here and we still have items left, process
	if len(queueItems) > 0 {
		createFileRows(db, processFilesAsync(queueItems))
	}

	return nil
}

func fileIsInDatabase(path string, db *gorm.DB) bool {
	md5 := HashStringMd5(path)

	HostName, err := os.Hostname()

	if err != nil {
		panic(err)
	}

	file := File{}

	db.Where(&File{FullPathMd5: md5, HostName: HostName}).First(&file)

	if len(file.FullPathMd5) > 0 {
		return true
	}

	return false
}

// Async function, takes a list of paths and turns it into a list of information needed for db insertion
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

			fmt.Printf("Processing %s\n", pi.Path)

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

// Gets all the info needed for insert into db
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
		FullPathMd5:        HashStringMd5(pi.Path),
		Path:               strings.ReplaceAll(pi.Path, conf.SearchDirectory, ""),
		Base:               conf.SearchDirectory,
		FileSizeBytes:      pi.Size,
		ExtensionLowerCase: trimLeftChars(strings.ToLower(filepath.Ext(pi.Path)), 1),
		Crc32:              Crc32,
		HostName:           HostName}

	return file, nil

}
