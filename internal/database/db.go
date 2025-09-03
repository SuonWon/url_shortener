package database

import (
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	schema "github.com/url_shortener/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectionDatabase() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")

	psqlSetup := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Bangkok", host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(psqlSetup), &gorm.Config{})

	if err != nil {
		fmt.Println("There is an error while connect to the database ", err)
	}
	// auto migreate code
	if err := AutoMigrate(db); err != nil {
		fmt.Println("Failed to migrate schema: ", err)
	}
	if db == nil {
		log.Fatal("db is nil til connection")
	} else {
		fmt.Println("Database connection is success!")
	}

	return db
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
		return err
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS citext`).Error; err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&schema.User{},
		&schema.CustomDomain{},
		&schema.ShortLink{},
		&schema.ShortLinkWithIndexes{},
	); err != nil {
		return err
	}

	return nil
}
