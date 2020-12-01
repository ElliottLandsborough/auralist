package main

import "time"

// File with music in it
type File struct {
	ID                 uint
	FullPathHash       uint32 `gorm:"index"` // murmur3(FullPath)
	FullPath           string // /home/ubuntu/Music/donk.mp3
	FileName           string // donk.mp3
	FileSizeBytes      int64  // file size in bytes (maximum 4294967295, 4gb!)
	ExtensionLowerCase string `gorm:"index"` // mp3
	Crc32              int64  `gorm:"index"` // 321789321
	Crc32WithoutTags   uint   `gorm:"index"` // 128291004 TODO
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ID3Tag parsed from mp3
type ID3Tag struct {
	ID        uint
	FileID    uint `gorm:"index"`
	Title     string
	Artist    string
	Album     string
	Year      string
	Genre     string
	Comments  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
