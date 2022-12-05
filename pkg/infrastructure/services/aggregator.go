package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
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

type HandlerAggregator struct {
	handlersMutex      sync.RWMutex
	registeredHandlers []application.Handler `inject:"mqtt.message_handler"`

	brokers  application.Repository[*entities.Broker] `inject:""`
	gateways application.GatewayRepository            `inject:""`

	clients sync.Map
}

func (a *HandlerAggregator) Register(handlers ...application.Handler) {
	a.handlersMutex.Lock()
	a.registeredHandlers = append(a.registeredHandlers, handlers...)
	a.handlersMutex.Unlock()
}

func (a *HandlerAggregator) Unregister(handler application.Handler) {
	for index, h := range a.registeredHandlers {
		if unsafe.SameInterfacePointer(handler, h) {
			a.handlersMutex.Lock()
			a.registeredHandlers = append(a.registeredHandlers[:index], a.registeredHandlers[index+1:]...)
			a.handlersMutex.Unlock()
		}
	}
}

func (a *HandlerAggregator) GetClient(clientId string) mqtt.Client {
	if client, ok := a.clients.Load(clientId); ok {
		return client.(mqtt.Client)
	}

	return nil
}

func (a *HandlerAggregator) DeleteClient(clientId string) {
	if client, ok := a.clients.LoadAndDelete(clientId); ok {
		client.(mqtt.Client).Disconnect(250)
	}
}

func (a *HandlerAggregator) Subscribe(topics map[string]byte, options *mqtt.ClientOptions) {
	client, ok := a.clients.LoadOrStore(options.ClientID, mqtt.NewClient(options))

	if !ok {
		token := client.(mqtt.Client).Connect()
		token.Wait()

		if token.Error() != nil {
			panic(token.Error())
		}
	}

	client.(mqtt.Client).SubscribeMultiple(topics, func(client mqtt.Client, message mqtt.Message) {
		a.handlersMutex.RLock()
		for _, handler := range a.registeredHandlers {
			if handler.CanHandle(client, message) {
				go handler.Handle(client, message)
			}
		}
		a.handlersMutex.RUnlock()
	})
}

func (a *HandlerAggregator) Unsubscribe(client mqtt.Client, topics map[string]byte) {
	for topic := range topics {
		client.Unsubscribe(topic)
	}
}

func (a *HandlerAggregator) Launch(context.Context) {
	for _, item := range a.brokers.FindAll() {
		if len(item.GetTopics()) > 0 {
			go a.Subscribe(item.GetTopics(), GetClientOptions(item))
		}
	}

	for _, item := range a.gateways.FindAll() {
		if len(item.GetTopics()) > 0 {
			go a.Subscribe(item.GetTopics(), GetClientOptions(item.Broker))
		}
	}
}

func (a *HandlerAggregator) Shutdown(context.Context) {
	var wait sync.WaitGroup
	a.clients.Range(func(_, client any) bool {
		wait.Add(1)
		go func() {
			client.(mqtt.Client).Disconnect(250)
			wait.Done()
		}()
		return true
	})

	wait.Wait()
}
