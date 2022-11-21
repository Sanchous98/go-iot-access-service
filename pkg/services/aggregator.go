package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"sync"
	"unsafe"
)

type Connectable interface {
	GetId() uuid.UUID
	GetOptions() *mqtt.ClientOptions
	GetTopics() map[string]byte
}

type Handler interface {
	Handle(client mqtt.Client, message mqtt.Message)
	CanHandle(client mqtt.Client, message mqtt.Message) bool
}

type HandlerFunc func(client mqtt.Client, message mqtt.Message)

func (h HandlerFunc) Handle(client mqtt.Client, message mqtt.Message) { h(client, message) }
func (h HandlerFunc) CanHandle(mqtt.Client, mqtt.Message) bool        { return true }

type Aggregatable interface {
	Connectable
	repositories.WithEndpoint
}

type HandlerAggregator[T Aggregatable] struct {
	registeredHandlers []Handler                  `inject:"mqtt.message_handler"`
	repository         repositories.Repository[T] `inject:""`
	mu                 sync.Mutex

	clients sync.Map
}

func (a *HandlerAggregator[T]) Register(handlers ...Handler) {
	a.mu.Lock()
	a.registeredHandlers = append(a.registeredHandlers, handlers...)
	a.mu.Unlock()
}

func (a *HandlerAggregator[T]) Unregister(handler Handler) {
	a.mu.Lock()
	for index, h := range a.registeredHandlers {
		if sameHandler(handler, h) {
			a.registeredHandlers = append(a.registeredHandlers[:index], a.registeredHandlers[index+1:]...)
		}
	}
	a.mu.Unlock()
}

func (a *HandlerAggregator[T]) Publish(model T, message []byte, qos byte) <-chan mqtt.Token {
	if client, ok := a.clients.Load(model.GetId()); ok {
		results := make(chan mqtt.Token, len(model.GetTopics()))
		for topic := range model.GetTopics() {
			results <- client.(mqtt.Client).Publish(topic, qos, false, message)
		}

		close(results)
		return results
	}

	return nil
}

func (a *HandlerAggregator[T]) Launch(ctx context.Context) {
	items := a.repository.FindAll()

	for _, item := range items {
		go func(item T) {
			client := a.Subscribe(item.GetTopics(), item.GetOptions(), func(client mqtt.Client, message mqtt.Message) {
				a.mu.Lock()
				for _, handler := range a.registeredHandlers {
					if handler.CanHandle(client, message) {
						handler.Handle(client, message)
					}
				}
				a.mu.Unlock()
			})

			a.clients.Store(item.GetId(), client)
		}(item)
	}
}

func (a *HandlerAggregator[T]) GetClient(model T) mqtt.Client {
	if client, ok := a.clients.Load(model.GetId()); ok {
		return client.(mqtt.Client)
	}

	return nil
}

func (a *HandlerAggregator[T]) Subscribe(topics map[string]byte, options *mqtt.ClientOptions, callback func(client mqtt.Client, message mqtt.Message)) mqtt.Client {
	client := mqtt.NewClient(options)
	token := client.Connect()
	token.Wait()
	client.SubscribeMultiple(topics, callback)
	return client
}

func (a *HandlerAggregator[T]) Unsubscribe(client mqtt.Client, topics map[string]byte) {
	for topic := range topics {
		client.Unsubscribe(topic)
	}
}

func (a *HandlerAggregator[T]) Shutdown(ctx context.Context) {
	a.clients.Range(func(_, client any) bool {
		go client.(mqtt.Client).Disconnect(250)
		return true
	})
}

func sameHandler(handler1, handler2 Handler) bool {
	return (*(*[2]uintptr)(unsafe.Pointer(&handler1)))[1] == (*(*[2]uintptr)(unsafe.Pointer(&handler2)))[1]
}
