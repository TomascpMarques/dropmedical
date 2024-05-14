package setup

import (
	"log"
	"os"

	database "github.com/TomascpMarques/dropmedical/database"
	http_api "github.com/TomascpMarques/dropmedical/http_api"
	models "github.com/TomascpMarques/dropmedical/models"
	gin "github.com/gin-gonic/gin"
	godotenv "github.com/joho/godotenv"
)

// app environments
const (
	PROD  = "production"
	LOCAL = "local"
)

func LoadEnvironment() {
	app_environment, is_defined := os.LookupEnv("ENVIRONMENT")
	if !is_defined {
		log.Fatalln("Variável de ambiente ENVIRONMENT não definida, por favor define a mesma")
	}

	var err error
	switch app_environment {
	case PROD:
		err = godotenv.Load("./.env/prod.env")
	case LOCAL:
		err = godotenv.Load("./.env/local.env")
	default:
		log.Fatalln("Valor para variavel de ambiente errado, não se encontra nos parametros aceitaveis")
	}

	if err != nil {
		log.Fatalf("Erro ao carregar variáveis de ambiente: %s", err.Error())
	}
}

func SetupGinApp() (engine *gin.Engine, err error) {
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Printf("DB Error: %s\n", err)
		os.Exit(1)
	}
	models.MigrateAll(db)

	engine = gin.Default()
	http_api.SetupRoutesGroup(engine, db)

	return
}
