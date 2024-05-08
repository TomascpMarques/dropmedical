package http_api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/TomascpMarques/dropmedical/database"
	"github.com/TomascpMarques/dropmedical/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var wg sync.WaitGroup

func TestCreateServer(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Failed to read the environment variables: %s\n", err)
	}

	r := gin.Default()

	db, _ := database.NewPostgresConnection()

	models.MigrateAll(db)

	SetupRoutesGroup(r, db)

	go func() {
		err := r.Run()
		if err != nil {
			wg.Done()
		}
	}()
	wg.Wait()
}

func TestShouldCreateDropper(t *testing.T) {
	wg.Add(1)

	resp, err := http.PostForm("http://localhost:8080/api/dropper", url.Values{
		"name":        {"Supa"},
		"machine_url": {"Something"},
		"active":      {"true"},
	})
	if err != nil {
		wg.Done()
		t.Fatalf("Error: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		wg.Done()
		t.Fatalf("Error!!!")
	}

	read, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	value := string(read)

	fmt.Printf("Value: %s\n", value)

	if len(read) < 1 {
		wg.Done()
		t.Fatalf("Body não válido")
	}

	wg.Done()
}
