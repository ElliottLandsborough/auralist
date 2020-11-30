package main

func main() {
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
