package setup

import (
	"log"
	"os"

	http_api "github.com/TomascpMarques/dropmedical/http_api"
	models "github.com/TomascpMarques/dropmedical/models"
	gin "github.com/gin-gonic/gin"
	godotenv "github.com/joho/godotenv"
	"gorm.io/gorm"
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

func SetupGinApp(db *gorm.DB, ch *chan models.MqttActionRequest) (engine *gin.Engine, err error) {
	models.MigrateAll(db)

	engine = gin.Default()
	http_api.SetupRoutesGroup(engine, db, ch)

	return
}
