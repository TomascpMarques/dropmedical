package setup

import (
	"log"

	database "github.com/TomascpMarques/dropmedical/database"
	http_api "github.com/TomascpMarques/dropmedical/http_api"
	gin "github.com/gin-gonic/gin"
	godotenv "github.com/joho/godotenv"
)

func SetupGinApp() (engine *gin.Engine, err error) {
	err = godotenv.Load(".env")
	if err != nil {
		log.Printf("Failed to read the environment variables: %s\n", err)
	}

	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Printf("Error: %s\n", err)
	}

	engine = gin.Default()
	http_api.SetupRoutesGroup(engine, db)

	return
}
