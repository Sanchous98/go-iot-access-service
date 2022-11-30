package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
	"unsafe"
)

type Connectable interface {
	GetOptions() *mqtt.ClientOptions
	GetTopics() map[string]byte
}

type Handler interface {
	Handle(client mqtt.Client, message mqtt.Message)
	CanHandle(client mqtt.Client, message mqtt.Message) bool
}

type Aggregatable interface {
	Connectable
	repositories.WithResource
}

type HandlerFunc func(client mqtt.Client, message mqtt.Message)

func (h HandlerFunc) Handle(client mqtt.Client, message mqtt.Message) { h(client, message) }
func (h HandlerFunc) CanHandle(mqtt.Client, mqtt.Message) bool        { return true }

type HandlerAggregator struct {
	handlersMutex      sync.RWMutex
	registeredHandlers []Handler `inject:"mqtt.message_handler"`

	brokers  repositories.Repository[*models.Broker]  `inject:""`
	gateways repositories.Repository[*models.Gateway] `inject:""`

	clients sync.Map
}

func (a *HandlerAggregator) Register(handlers ...Handler) {
	a.handlersMutex.Lock()
	a.registeredHandlers = append(a.registeredHandlers, handlers...)
	a.handlersMutex.Unlock()
}

func (a *HandlerAggregator) Unregister(handler Handler) {
	for index, h := range a.registeredHandlers {
		if sameHandler(handler, h) {
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
			go a.Subscribe(item.GetTopics(), item.GetOptions())
		}
	}
	for _, item := range a.gateways.FindAll() {
		if len(item.GetTopics()) > 0 {
			go a.Subscribe(item.GetTopics(), item.GetOptions())
		}
	}
}

func (a *HandlerAggregator) Shutdown(context.Context) {
	a.clients.Range(func(_, client any) bool {
		go client.(mqtt.Client).Disconnect(250)
		return true
	})
}

func sameHandler(handler1, handler2 Handler) bool {
	return (*(*[2]uintptr)(unsafe.Pointer(&handler1)))[1] == (*(*[2]uintptr)(unsafe.Pointer(&handler2)))[1]
}
