package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	tag "github.com/dhowden/tag"
	"gorm.io/gorm"
)

// Tag parsed from mp3
type Tag struct {
	ID        uint
	FileID    uint `gorm:"index"`
	Title     string
	Artist    string
	Album     string
	Year      string
	Genre     string
	Sum       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func parseTagsToDb(file File, db *gorm.DB) {
	f, err := os.Open(file.Path)

	if err != nil {
		panic(err) // deal with this soon...
	}

	sum, err := tag.Sum(f)

	if err != nil {
		panic(err) // deal with this soon...
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
