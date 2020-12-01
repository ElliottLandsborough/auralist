package main

import (
	id3 "github.com/mikkyang/id3-go"
	"gorm.io/gorm"
)

func parseID3TagsToDb(file File, db *gorm.DB) {
	id3File, err := id3.Open(file.FullPath)

	if err != nil {
		panic(err)
	}

	db.Create(&ID3Tag{
		FileID: file.ID,
		Title:  id3File.Title(),
		Artist: id3File.Artist(),
		Album:  id3File.Album(),
		Year:   id3File.Year(),
		Genre:  id3File.Genre(),
		//Comments: id3File.Comments()}) lolcba
	})
}
