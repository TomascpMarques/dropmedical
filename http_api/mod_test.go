package http_api

import (
	"bytes"
	"encoding/json"
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
	err := godotenv.Load("../.env/local.env")
	if err != nil {
		log.Fatalf("X Failed to read the environment variables: %s\n", err)
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

	resp := createDropper(t, "supa", "none")

	fmt.Printf("Value: %v\n", resp)

	wg.Done()
}

func createDropper(t *testing.T, name, machine_url string) (dropper models.Dropper) {
	resp, err := http.PostForm("http://localhost:8080/api/dropper", url.Values{
		"name":        {name},
		"machine_url": {machine_url},
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
	value := []byte(read)

	err = json.Unmarshal(value, &dropper)
	if err != nil {
		t.Fatalf("Erro: %s", err.Error())
	}
	return
}

func TestShouldReloadDropperSection(t *testing.T) {
	wg.Add(1)
	defer wg.Done()

	dropper := createDropper(t, "supper", "yes")

	sectionReload := reloadDropperSection{
		Dropper:  dropper.ID,
		Section:  1,
		PillName: "Brufen",
		Quantity: 3,
	}

	json, _ := json.Marshal(sectionReload)

	resp, err := http.Post(
		"http://localhost:8080/api/dropper/section/reload",
		"application/json",
		bytes.NewBuffer(json),
	)
	if err != nil {
		t.Fatalf("Erro: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Fatalf("Erro!!!")
	}
}
