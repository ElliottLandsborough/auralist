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
		case "parsetags":
			parseTags()
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

	// iterate through files
	listFiles(db)
}

func parseTags() {
	// check db is ready
	db, e := getDB()

	if e != nil {
		panic(e)
	}

	deleteAllTags(db)

	// migrate
	db.AutoMigrate(&File{})
	db.AutoMigrate(&Tag{})

	var file File

	rows, e := db.Model(&File{}).Rows()

	if e != nil {
		panic(e)
	}

	for rows.Next() {
		db.ScanRows(rows, &file)
		parseTagsToDb(file, db)
	}
}
