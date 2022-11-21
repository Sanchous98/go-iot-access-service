package handlers

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
)

type LogHandler struct {
	appEnv string `env:"APP_ENV"`
}

func (h *LogHandler) Handle(_ mqtt.Client, message mqtt.Message) {
	log.Println(utils.BytesToStr(message.Payload()))
}

func (h *LogHandler) CanHandle(mqtt.Client, mqtt.Message) bool {
	return h.appEnv == "dev" || h.appEnv == "development"
}
