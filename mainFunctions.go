package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Initialization routine.
func init() {
	// Retrieve config options.
	conf = getConf()
}

func listFiles() ([]string, error) {
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
