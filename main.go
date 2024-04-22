// Main entry for the API
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/TomascpMarques/dropmedical/api"
	"github.com/TomascpMarques/dropmedical/database"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Failed to read the environment variables: %s", err)
	}

	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	r := gin.Default()

	// db, _ := controllers.NewDataSource(controllers.Memory)

	api.SetupRoutesGroup(r, db)

	_ = r.Run() // listen and serve on 0.0.0.0:8080
}
