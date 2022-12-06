package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/unsafe"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
)

type Connectable interface {
	GetTopics() map[string]byte
}

type Aggregatable interface {
	Connectable
	application.WithResource
}

type MessageAggregator struct {
	ClientsPool

	handlersMutex      sync.RWMutex
	registeredHandlers []application.Handler `inject:"mqtt.message_handler"`

	brokers  application.Repository[*entities.Broker] `inject:""`
	gateways application.GatewayRepository            `inject:""`
	log      logger.Logger                            `inject:""`
}

func (a *MessageAggregator) Register(handlers ...application.Handler) {
	a.handlersMutex.Lock()
	a.registeredHandlers = append(a.registeredHandlers, handlers...)
	a.handlersMutex.Unlock()
}

func (a *MessageAggregator) Unregister(handler application.Handler) {
	for index, h := range a.registeredHandlers {
		if unsafe.SameInterfacePointer(handler, h) {
			a.handlersMutex.Lock()
			a.registeredHandlers = append(a.registeredHandlers[:index], a.registeredHandlers[index+1:]...)
			a.handlersMutex.Unlock()
		}
	}
}

func (a *MessageAggregator) Subscribe(topics map[string]byte, options *mqtt.ClientOptions) {
	client := a.CreateClient(options)

	if !client.IsConnected() {
		token := client.Connect()
		token.Wait()

		if err := token.Error(); err != nil {
			a.log.Errorln(err)
			return
		}
	}

	client.SubscribeMultiple(topics, func(client mqtt.Client, message mqtt.Message) {
		a.handlersMutex.RLock()
		for _, handler := range a.registeredHandlers {
			if handler.CanHandle(client, message) {
				go handler.Handle(client, message)
			}
		}
		a.handlersMutex.RUnlock()
	})
}

func (a *MessageAggregator) Unsubscribe(client mqtt.Client, topics map[string]byte) {
	for topic := range topics {
		client.Unsubscribe(topic)
	}
}

func (a *MessageAggregator) Launch(context.Context) {
	for _, item := range a.brokers.FindAll() {
		if len(item.GetTopics()) > 0 {
			go a.Subscribe(item.GetTopics(), application.GetClientOptions(a.log, item))
		}
	}

	for _, item := range a.gateways.FindAll() {
		if len(item.GetTopics()) > 0 {
			go a.Subscribe(item.GetTopics(), application.GetClientOptions(a.log, item.Broker))
		}
	}
}

func (a *MessageAggregator) Shutdown(context.Context) {
	a.ClientsPool.Purge()
}
