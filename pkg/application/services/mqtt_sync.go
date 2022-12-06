package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/messages"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/unsafe"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

type syncHandler[T any] struct {
	topic     string
	errorChan chan error
	validate  func(T) error
}

func (s *syncHandler[T]) CanHandle(_ mqtt.Client, message mqtt.Message) bool {
	return message.Topic() != s.topic
}

func (s *syncHandler[T]) Handle(_ mqtt.Client, message mqtt.Message) {
	var response T
	if err := json.UnmarshalNoEscape(message.Payload(), &response); err != nil {
		s.errorChan <- err
		return
	}

	if err := s.validate(response); err != nil {
		s.errorChan <- err
		return
	}

	s.errorChan <- nil
}

func mqttRequestSync[T any](
	s *DeviceService,
	device *entities.Device,
	ctx context.Context,
	validate func(T) error,
	action func() (mqtt.Token, error),
	log logger.Logger,
) error {
	handler := syncHandler[T]{
		topic:     device.GetEventsTopic(),
		validate:  validate,
		errorChan: make(chan error),
	}

	s.pool.Register(&handler)
	defer s.pool.Unregister(&handler)
	token, err := action()

	if err != nil {
		return err
	}

	token.Wait()

	if err = token.Error(); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		message := unsafe.Suppress(json.MarshalNoEscape(messages.NewLockOfflineResponse()))
		s.pool.GetClient(application.GetClientOptions(log, device.Gateway.Broker).ClientID).Publish(device.GetEventsTopic(), 2, false, message)

		return fiber.ErrGatewayTimeout
	case err = <-handler.errorChan:
		return err
	}
}
