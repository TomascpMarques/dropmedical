package mqtt_api

import (
	"log"
	"strings"
	"time"

	"github.com/wind-c/comqtt/v2/mqtt"
	"github.com/wind-c/comqtt/v2/mqtt/hooks/auth"
	"github.com/wind-c/comqtt/v2/mqtt/packets"
)

// MQTT Topics
const (
	HealthCheckROOT = "health"
	HealthCheckIsUp = "/up"
	// -----------------------
	DevicesROOT     = "devices/disp"
	DevicesRegister = "/register/#"
)

// NewMqttServer cria um novo servidor mqtt que permite comunicação full duplex
func NewMqttServer() *mqtt.Server {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true, // you must enable inline client to use direct publishing and subscribing.
	})

	server.AddHook(new(auth.AllowHook), nil)

	err := server.Subscribe(DevicesROOT+DevicesRegister, 1, func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		topic := strings.Split(pk.TopicName, "/")
		device_id := topic[len(topic)-1]

		if device_id == "" {
			server.Log.Info(DevicesROOT+DevicesRegister, "status", "responded")

			err := cl.WritePacket(packets.Packet{
				FixedHeader: packets.FixedHeader{
					Type: packets.Publish,
				},
				TopicName: pk.TopicName,
				Payload:   []byte("AAAAAA"),
			})
			if err != nil {
				log.Println(err.Error())
			}
		}
	})

	if err != nil {
		log.Fatalln("Falha ao atribuir subscriber MqTT para o registo de dispensers")
	}

	// Server health check
	go func() {
		for {
			server.Log.Info("server.health_check", "status", "sent")
			server.PublishToSubscribers(packets.Packet{
				FixedHeader: packets.FixedHeader{
					Type: packets.Publish,
				},
				TopicName: HealthCheckROOT + HealthCheckIsUp,
				Payload:   []byte("UP"),
			}, true)
			time.Sleep(time.Minute / 2)
		}
	}()

	return server
}
