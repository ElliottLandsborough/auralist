package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Starting...\n")

	if len(os.Args[1:]) > 0 {

		arg := os.Args[1]

		switch arg {
		case "collectPaths":
			collectPaths()
		case "processPaths":
			processPaths()
		case "parsetags":
			parseTags()
		case "listen":
			server()
		default:
			fmt.Printf("Please choose a command.\n")
		}
	}

	fmt.Printf("Finished.\n")
}

func collectPaths() {
	fileName := "cache/paths.log"

	// Versioning??
	//unixTime := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	//fileName := "cache/paths." + unixTime + ".log"

	readPaths(fileName)
}

func processPaths() {
	fileName := "cache/paths.log"

	// check db is ready
	db, e := getDB()

	if e != nil {
		panic(e)
	}

	// migrate
	db.AutoMigrate(&File{})

	handlePaths(getFilePaths(fileName), db)
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

/**
  42856 mp3
   6340 flac
   5768 jpg
   3121 m3u
   2884 nfo
   2753 sfv
    917 wav
    713 m4a
    513 ini
    454 cue
    343 wma
    338 jpeg
    323 asd
    314 itc2
    292 db
    286 log
    119 txt
    103 pkf
    100 ogg
     97 mpc
     88 aif
     86 message
     83 png
     48 ds_store
     26 sfk
     23 gif
     23 bmp
     19 vob
     19 avi
     17 rar
     16 url
     13 pls
     13 pdf
     13 m3u8
     11 html
     11 doc
     10 als
      9 mp4
      8 accurip
      7 ifo
      7 dat
      7 bup
      6 ico
      5 rtf
      5 pk
      5 mxm
      5 lnk
      5 cfg
      5 aiff
      4 zip
      4 md5
      4 !ut
      3 xmp
      3 xml
      3 fla
      3 1
      2 xxx
      2 tqd
      2 mpg
      2 mdd
      2 itdb
      2 dcr
      2 aucdtect
      2 alt
      2 aac
      1 xm
      1 rm
      1 mov
      1 mkv
      1 mid
**/
