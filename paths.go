package main

import (
	"bufio"
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
func readPaths(cacheFile string) {
	if _, err := os.Stat(cacheFile); err == nil {
		panic("Clear your cache first?")
	}

	log.Println("Collecting paths...")

	e := filepath.Walk(conf.SearchDirectory, func(path string, f os.FileInfo, err error) error {
		// Don't process directories
		if !f.IsDir() {
			writePathToCache(path, cacheFile)
		}

		return err
	})

	if e != nil {
		panic(e)
	}
}

// Write a path line to the cache file
func writePathToCache(path string, cacheFile string) {
	f, err := os.OpenFile(cacheFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	if _, err := f.WriteString(path + "\n"); err != nil {
		log.Println(err)
	}
}

func getFilePaths(fi string) []string {
	if _, err := os.Stat(fi); os.IsNotExist(err) {
		panic("Paths have not been collected yet")
	}

	paths := make([]string, 0)

	f, err := os.Open(fi)
	if err != nil {
		fmt.Println("error opening file= ", err)
		os.Exit(1)
	}
	r := bufio.NewReader(f)
	s, e := Readln(r)
	for e == nil {
		// If the path is not zero length
		if len(s) > 0 {
			paths = append(paths, s)
		}

		// Keep going until lines run out
		s, e = Readln(r)
	}

	return paths
}
