// Main entry for the API
package main

import (
	"log"
	"net"
	"sync"

	"github.com/wind-c/comqtt/v2/mqtt/listeners"

	"github.com/TomascpMarques/dropmedical/mqtt_api"
	"github.com/TomascpMarques/dropmedical/setup"
)

var wg sync.WaitGroup

func main() {
	// Inicialização do servidor de MQTT
	server := mqtt_api.NewMqttServer()
	tcp_listener_mqtt := listeners.NewTCP("t1", ":1883", nil)

	err := server.AddListener(tcp_listener_mqtt)
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
	}()
	// ---------------------------------

	// Inicialização do servidor web
	wg.Add(1)
	tcp_listener_gin, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		app, err := setup.SetupGinApp() // listen and serve on 0.0.0.0:8081 / [::]:8081
		if err != nil {
			log.Fatal(err)
			wg.Done()
		}
		_ = app.RunListener(tcp_listener_gin)
	}()
	// ----------------------------------

	wg.Wait()
}
