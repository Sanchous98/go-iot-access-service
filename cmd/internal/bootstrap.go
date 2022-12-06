package internal

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	loggerWrapper "bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/logger"
	"github.com/Sanchous98/go-di"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

func Bootstrap(container di.GlobalState) {
	mqtt.ERROR = loggerWrapper.NewMqttErrorWrapper(container.Get(new(logger.Logger)).(logger.Logger))
	mqtt.WARN = loggerWrapper.NewMqttWarnWrapper(container.Get(new(logger.Logger)).(logger.Logger))
	mqtt.CRITICAL = loggerWrapper.NewMqttCriticalWrapper(container.Get(new(logger.Logger)).(logger.Logger))
	//mqtt.DEBUG = loggerWrapper.NewMqttDebugWrapper(container.Get(new(logger.Logger)).(logger.Logger))

	profiler := fiber.New(fiber.Config{
		JSONDecoder: func(data []byte, v any) error { return json.UnmarshalNoEscape(data, v) },
		JSONEncoder: json.MarshalNoEscape,
	})
	profiler.Use(pprof.New())
	container.Get(new(logger.Logger)).(logger.Logger).Errorln(profiler.Listen(":6060"))
}
