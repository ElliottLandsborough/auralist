package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Gets db connection info
func getDSN() string {
	MysqlDatabase := conf.MysqlDatabase
	MysqlUser := conf.MysqlUser
	MysqlPass := conf.MysqlPass
	MysqlHost := conf.MysqlHost

	return fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8&parseTime=True&loc=Local",
		MysqlUser,
		MysqlPass,
		MysqlHost,
		MysqlDatabase)
}

// Gets a db connection
func getDB() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(getDSN()), &gorm.Config{})

	return db, err
}
