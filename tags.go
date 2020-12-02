package main

import (
	"fmt"
	"os"
	"strconv"

	tag "github.com/dhowden/tag"
	"gorm.io/gorm"
)

func parseTagsToDb(file File, db *gorm.DB) {
	f, err := os.Open(file.FullPath)

	if err != nil {
		panic(err)
	}

	sum, err := tag.Sum(f)

	if err != nil {
		panic(err)
	}

	// try tag read method 1
	m, err := tag.ReadFrom(f)

	if err == nil {
		year := ""

		if m.Year() > 0 {
			year = strconv.Itoa(m.Year())
		}

		fmt.Printf("%s - %s - %s\n", m.Artist(), m.Title(), m.Album())

		db.Create(&Tag{
			FileID: file.ID,
			Title:  m.Title(),
			Artist: m.Artist(),
			Album:  m.Album(),
			Year:   year,
			Genre:  m.Genre(),
			Sum:    sum})
	}
}
