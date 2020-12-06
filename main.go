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
		case "syncFiles":
			syncFiles()
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

func syncFiles() {
	// check db is ready
	db, e := getDB()

	if e != nil {
		panic(e)
	}

	// Connect to server
	sshClient := getSSHClient()

	// Get local hostname
	localHostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	// Get remote hostname
	remoteHostName, err := remoteRun("hostname", getSSHSession(sshClient))

	// Empty array of files
	files := make([]File, 0)

	// Get all files for this hostname
	db.Where(&File{HostName: localHostName}).Find(&files)

	// Loop through all the files
	for _, file := range files {
		// Empty array of files
		potentialDuplicate := File{}

		// Check remote location for file

		// Check db for first result where crc32 match
		db.Where(&File{HostName: remoteHostName, Crc32: file.Crc32}).First(&potentialDuplicate)

		// If file size is zero
		// manually generate it
		// stop loop here

		// If we got a match
		if len(potentialDuplicate.FileName) > 0 {
			// Does local md5 match remote md5?
		}

		// Copy old file to new location with all

		// If current local file doesn't have a duplicate on remote server

		// Upload it
	}

	/*
			remotePath := conf.RemotePath

			scpClient, err := scp.NewClientBySSH(sshClient)
			if err != nil {
				fmt.Println("Error creating new SSH session from existing connection", err)
			}

			// Open a file
			f, _ := os.Open("/proc/cpuinfo")

			// Close client connection after the file has been copied
			defer scpClient.Close()

			// Close the file after it has been copied
			defer f.Close()

			remoteFilePath := remotePath + "cpuinfo"

			// Usage: CopyFile(fileReader, remotePath, permission)
			err = scpClient.CopyFile(f, remoteFilePath, "0644")

			// Get md5 of remote file
			md5sum := hashFileMD5Remote(remoteFilePath, sshClient)

			// Get sha1 of remote file
			sha1sum := hashFileSHA1Remote(remoteFilePath, sshClient)

			if err != nil {
				panic(err)
			}

			fmt.Println(hostName)
			fmt.Println(md5sum)
			fmt.Println(sha1sum)

			if err != nil {
				fmt.Println("Error while copying file ", err)
		  }
	*/
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
