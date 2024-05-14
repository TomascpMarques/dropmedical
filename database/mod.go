// Package database - Handles database data sources
package database

import (
	"log"
	"os"
	"time"

	postgres "gorm.io/driver/postgres"
	gorm "gorm.io/gorm"
)

// NewPostgresConnection creates a new postgres db connection
func NewPostgresConnection() (db *gorm.DB, err error) {
	postgresDSN, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatalf("Variável de ambiente para URL da base de dados não encontrada")
	}

	for range 5 {
		log.Printf("Trying...")
		db, err = gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(time.Second * 3)
	}
	return db, err
}
