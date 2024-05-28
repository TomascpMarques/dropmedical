// Main entry for the API
package main

import (
	"log"
	"net"
	"os"
	"sync"

	listeners "github.com/wind-c/comqtt/v2/mqtt/listeners"

	database "github.com/TomascpMarques/dropmedical/database"
	models "github.com/TomascpMarques/dropmedical/models"
	mqtt_api "github.com/TomascpMarques/dropmedical/mqtt_api"
	setup "github.com/TomascpMarques/dropmedical/setup"
)

var wg sync.WaitGroup
var interop_mqtt_channel chan models.MqttActionRequest = make(chan models.MqttActionRequest, 20)

func main() {
	// Load env files
	setup.LoadEnvironment()

	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Printf("DB Error: %s\n", err)
		os.Exit(1)
	}

	// Inicialização do servidor de MQTT
	server := mqtt_api.NewMqttServer()
	tcp_listener_mqtt := listeners.NewTCP("tcp_mqtt_1", ":1883", nil)

	err = server.AddListener(tcp_listener_mqtt)
	if err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
			wg.Done()
		}
		// Loop for messages
		for {
			msg, r := <-interop_mqtt_channel
			if r {
				// Publish
				server.Publish(msg.Topic, []byte(msg.Value), false, 0)
			}
		}
	}()
	// ---------------------------------

	// Inicialização do servidor web
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
		log.Println("Defaulting to port " + port)
	}

	wg.Add(1)
	tcp_listener_gin, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	app, err := setup.SetupGinApp(db) // listen and serve on 0.0.0.0:port / [::]:port
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err = app.RunListener(tcp_listener_gin)
		if err != nil {
			log.Fatal(err)
			wg.Done()
		}
	}()
	// ----------------------------------

	// Start do cronjob
	wg.Add(1)
	go func() {
		for {
			if err := models.PillDispenseBGJob(db, interop_mqtt_channel); err != nil {
				log.Fatalf("Erro de cronjob: %+e", err)
				break
			}
		}
		wg.Done()
	}()
	// ----------------------------------

	wg.Wait()
}
