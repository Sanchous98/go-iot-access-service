package main

import (
	"bitbucket.org/4suites/iot-service-golang/messages"
	"bitbucket.org/4suites/iot-service-golang/models"
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"github.com/Sanchous98/go-di"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"log"
)

func main() {
	app := di.Application()
	app.Set(func(container di.Container) repositories.Repository[*models.Gateway] {
		return container.Build(new(repositories.GatewayRepository)).(*repositories.GatewayRepository)
	})

	app.Set(func(container di.Container) repositories.Repository[*models.Broker] {
		return container.Build(new(repositories.BrokerRepository)).(*repositories.BrokerRepository)
	})
	app.Set(new(repositories.DeviceRepository))
	app.LoadEnv()
	app.Compile()

	repository := app.Get((*repositories.DeviceRepository)(nil)).(*repositories.DeviceRepository)
	device := repository.FindByMacId("0x0000000001")
	client := mqtt.NewClient(device.GetGateway().GetOptions())
	token := client.Connect()
	token.Wait()
	client.Subscribe(device.GetCommandsTopic(), 2, func(client mqtt.Client, message mqtt.Message) {
		var request messages.EventRequest[messages.LockAuto]
		err := json.UnmarshalNoEscape(message.Payload(), &request)

		if err != nil {
			return
		}

		client.Publish(device.GetEventsTopic(), 0, false, messages.NewLockOfflineResponse())
		log.Println("responded")
	})

	select {}
}
