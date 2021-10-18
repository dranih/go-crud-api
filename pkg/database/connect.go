package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("../../test/test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	log.Printf("Connected to %s", db.Config.Name())
	return db
}
