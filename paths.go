package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// PathInfo contains a full path and other basic info
type PathInfo struct {
	Path string // /home/ubuntu/Music/donk.mp3
	Size int64  // file size in bytes (maximum 4294967295, 4gb!)
}

// Get all paths from folder specified in config
func readPaths() {
	paths := make([]string, 0)

	fmt.Println("Collecting paths...")

	fileName := "cache/paths.log"

	// Versioning??
	//unixTime := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	//fileName := "cache/paths." + unixTime + ".log"

	e := filepath.Walk(conf.SearchDirectory, func(path string, f os.FileInfo, err error) error {
		// Don't process directories
		if !f.IsDir() {
			writePathToCache(path, fileName)
		}

		return err
	})

	if e != nil {
		panic(e)
	}

	fmt.Printf("Found %d files.\n", len(paths))
}

// Write a path line to the cache file
func writePathToCache(path string, fileName string) {
	f, err := os.OpenFile(fileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	if _, err := f.WriteString(path + "\n"); err != nil {
		log.Println(err)
	}
}
