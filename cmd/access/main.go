package main

import (
	"bitbucket.org/4suites/iot-service-golang/api"
	"bitbucket.org/4suites/iot-service-golang/handlers"
	"bitbucket.org/4suites/iot-service-golang/models"
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"bitbucket.org/4suites/iot-service-golang/services"
	"github.com/Sanchous98/go-di"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"log"
	"os"
)

func main() {
	app := di.Application()
	app.AddEntryPoint(bootstrap)

	app.Set(new(api.Handler))
	app.Set(func(container di.Container) repositories.Repository[*models.Gateway] {
		return container.Build(new(repositories.GatewayRepository)).(*repositories.GatewayRepository)
	})

	app.Set(func(container di.Container) repositories.Repository[*models.Broker] {
		return container.Build(new(repositories.BrokerRepository)).(*repositories.BrokerRepository)
	})

	app.Set(new(services.HandlerAggregator[*models.Broker]))
	app.Set(new(services.HandlerAggregator[*models.Gateway]))
	app.Set(new(handlers.VerifyOnlineHandler), "mqtt.message_handler")
	//app.Set(new(handlers.LogHandler), "mqtt.message_handler")
	app.Run(true)
}

func bootstrap(container di.GlobalState) {
	if container.GetParam("APP_ENV") != "prod" && container.GetParam("APP_ENV") != "production" {
		mqtt.ERROR = log.New(os.Stdout, "[mqtt:ERROR]::", log.LUTC)
		mqtt.WARN = log.New(os.Stdout, "[mqtt:WARN]::", log.LUTC)
	}

	mqtt.CRITICAL = log.New(os.Stdout, "[mqtt:CRITICAL]::", log.LUTC)

	//if container.GetParam("APP_ENV") == "dev" || container.GetParam("APP_ENV") == "development" {
	//	mqtt.DEBUG = log.New(os.Stdout, "[mqtt:DEBUG]::", log.LUTC)
	//}

	profiler := fiber.New()
	profiler.Use(pprof.New())
	log.Println(profiler.Listen(":6060"))
}
