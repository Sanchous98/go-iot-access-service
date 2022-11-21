package main

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/api"
	"bitbucket.org/4suites/iot-service-golang/pkg/handlers"
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"context"
	"github.com/Sanchous98/go-di"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"log"
	"os"
)

func main() {
	for {
		launch()
	}
}

func launch() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if err := recover(); err != nil {
			log.Printf("%v", err)
			cancel()
		}
	}()

	app := di.Application(ctx)
	app.AddEntryPoint(bootstrap)

	app.Set(new(api.ServerApi))
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
	app.Run(app.LoadEnv)
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

	profiler := fiber.New(fiber.Config{
		JSONDecoder: func(data []byte, v any) error { return json.UnmarshalNoEscape(data, v) },
		JSONEncoder: func(v any) ([]byte, error) { return json.MarshalNoEscape(v) },
	})
	profiler.Use(pprof.New())
	log.Println(profiler.Listen(":6060"))
}
