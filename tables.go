package main

import "time"

// File with music in it
type File struct {
	ID                 uint
	FileName           string // donk.mp3
	Path               string // /home/ubuntu/Music/donk.mp3
	Base               string
	PathHash           uint32 `gorm:"index"` // murmur3(Path)
	FileSizeBytes      int64  // file size in bytes (maximum 4294967295, 4gb!)
	ExtensionLowerCase string `gorm:"index"` // mp3
	Crc32              int64  `gorm:"index"` // 321789321
	Md5                string `gorm:"index;size:32"`
	HostName           string `gorm:"index;size:256"` // max linux hostname size as per manpage
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

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
