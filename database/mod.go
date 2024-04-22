// Package database - Handles database data sources
package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewPostgresConnection creates a new postgres db connection
func NewPostgresConnection() (*gorm.DB, error) {
	postgresDSN, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatalf("Variavél de ambiente para URL da base de dados não encontrada")
	}

	return gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
}
