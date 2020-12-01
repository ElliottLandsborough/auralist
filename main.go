package main

import (
	"fmt"
	"os"
)

func main() {
	output := "auralist"

	if len(os.Args[1:]) > 0 {

		arg := os.Args[1]

		switch arg {
		case "index":
			indexAllFiles()
		case "parse:mp3":
			parseMp3()
		default:
			fmt.Printf("Please choose a command.")
		}
	}

	fmt.Printf(output)
}

func indexAllFiles() {
	// check db is ready
	db, e := getDB()

	if e != nil {
		panic(e)
	}

	deleteAllFiles(db)

	// migrate
	db.AutoMigrate(&File{})
	db.AutoMigrate(&ID3Tag{})

	// iterate through files
	listFiles(db)
}

func parseMp3() {
	// check db is ready
	db, e := getDB()

	if e != nil {
		panic(e)
	}

	// migrate
	db.AutoMigrate(&ID3Tag{})

	var file File

	rows, e := db.Model(&File{}).Rows()

	if e != nil {
		panic(e)
	}

	for rows.Next() {
		db.ScanRows(rows, &file)
		parseID3TagsToDb(file, db)
	}
}
