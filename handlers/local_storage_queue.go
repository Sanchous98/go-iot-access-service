package handlers

import (
	"bitbucket.org/4suites/iot-service-golang/messages"
	"bitbucket.org/4suites/iot-service-golang/models"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
)

type LocalStorageQueue struct {
	deviceQueues map[*models.Device][]messages.EventRequest[messages.LocalStorageEvent]
}

func (l *LocalStorageQueue) Handle(client mqtt.Client, message mqtt.Message) {
	var p messages.EventRequest[messages.LocalStorageEvent]
	_ = json.UnmarshalNoEscape(message.Payload(), &p)

	var gatewayIeee, deviceMacId string
	fmt.Sscanf(message.Topic(), "$foursuites/gw/%s/dev/%s/actions", gatewayIeee, deviceMacId)

}

func (l *LocalStorageQueue) CanHandle(client mqtt.Client, message mqtt.Message) bool {
	var p messages.EventRequest[messages.LocalStorageEvent]
	err := json.UnmarshalNoEscape(message.Payload(), &p)
	return err == nil
}
