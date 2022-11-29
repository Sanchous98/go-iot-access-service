package internal

import (
	"github.com/Sanchous98/go-di"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"log"
	"os"
)

func Bootstrap(container di.GlobalState) {
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
