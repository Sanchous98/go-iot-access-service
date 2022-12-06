package application

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type ClientPool interface {
	GetClient(clientId string) mqtt.Client
	DeleteClient(clientId string)
	Register(handlers ...Handler)
	Unregister(handler Handler)
}

type Handler interface {
	Handle(client mqtt.Client, message mqtt.Message)
	CanHandle(client mqtt.Client, message mqtt.Message) bool
}

type WithResource interface {
	GetResource() string
}

type Repository[T WithResource] interface {
	Find(id uuid.UUID) T
	FindAll() []T
	FindBy(map[string]any) []T
	FindOneBy(map[string]any) T
}

type GatewayRepository interface {
	Repository[*entities.Gateway]
	FindByIeee(string) *entities.Gateway
}

type DeviceRepository interface {
	Repository[*entities.Device]
	FindByMacId(string) *entities.Device
	FindByMacIdAndGatewayIeee(string, string) *entities.Device
}

type HandlerFunc func(client mqtt.Client, message mqtt.Message)

func (h HandlerFunc) Handle(client mqtt.Client, message mqtt.Message) { h(client, message) }
func (h HandlerFunc) CanHandle(mqtt.Client, mqtt.Message) bool        { return true }
