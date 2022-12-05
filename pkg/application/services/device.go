package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/messages"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/services"
	"context"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

type DeviceService struct {
	pool application.HandlerPool `inject:""`
}

func (s *DeviceService) Open(device *entities.Device, channels []int) (mqtt.Token, error) {
	message, _ := json.MarshalNoEscape(messages.NewLockOpenEvent(0, channels))

	client := s.pool.GetClient(services.GetClientOptions(device.Gateway.Broker).ClientID)

	if client == nil {
		log.Printf("Client %s does not exist\n", services.GetClientOptions(device.Gateway.Broker).ClientID)
		return nil, fiber.ErrInternalServerError
	}

	return client.Publish(device.GetCommandsTopic(), 2, false, message), nil
}
func (s *DeviceService) OpenSync(parent context.Context, device *entities.Device, channels []int) error {
	responseValidator := func(response messages.Response[messages.LockResponse]) error {
		if response.Payload.LockActionStatus == messages.LockOpenedLockStatus ||
			response.Payload.LockActionStatus == messages.ErrorLockAlreadyOpenLockStatus &&
				len(response.Payload.ChannelIds) > 0 {
			return nil
		}

		return errors.New("Command failed with status: " + string(response.Payload.LockActionStatus))
	}
	ctx, cancel := context.WithTimeout(parent, 7*time.Second)
	defer cancel()

	return mqttRequestSync[messages.Response[messages.LockResponse]](s, device, ctx, responseValidator, func() (mqtt.Token, error) { return s.Open(device, channels) })
}

func (s *DeviceService) Close(device *entities.Device) (mqtt.Token, error) {
	message, _ := json.MarshalNoEscape(messages.NewLockCloseEvent(0))

	client := s.pool.GetClient(services.GetClientOptions(device.Gateway.Broker).ClientID)

	if client == nil {
		log.Printf("Client %s does not exist\n", services.GetClientOptions(device.Gateway.Broker).ClientID)
		return nil, fiber.ErrInternalServerError
	}

	return client.Publish(device.GetCommandsTopic(), 2, false, message), nil
}
func (s *DeviceService) CloseSync(parent context.Context, device *entities.Device) error {
	responseValidator := func(response messages.Response[messages.LockResponse]) error {
		if response.Payload.LockActionStatus == messages.LockClosedLockStatus ||
			response.Payload.LockActionStatus == messages.ErrorLockAlreadyClosedLockStatus &&
				len(response.Payload.ChannelIds) > 0 {
			return nil
		}

		return errors.New("Command failed with status: " + string(response.Payload.LockActionStatus))
	}
	ctx, cancel := context.WithTimeout(parent, 7*time.Second)
	defer cancel()

	return mqttRequestSync[messages.Response[messages.LockResponse]](s, device, ctx, responseValidator, func() (mqtt.Token, error) { return s.Close(device) })
}

func (s *DeviceService) Auto(device *entities.Device, recloseDelay byte, channels []int) (mqtt.Token, error) {
	message, _ := json.Marshal(messages.NewLockAutoEvent(0, recloseDelay, channels))

	client := s.pool.GetClient(services.GetClientOptions(device.Gateway.Broker).ClientID)

	if client == nil {
		log.Printf("Client %s does not exist\n", services.GetClientOptions(device.Gateway.Broker).ClientID)
		return nil, fiber.ErrInternalServerError
	}

	return client.Publish(device.GetCommandsTopic(), 2, false, message), nil
}
func (s *DeviceService) AutoSync(parent context.Context, device *entities.Device, recloseDelay byte, channels []int) error {
	responseValidator := func(response messages.Response[messages.LockResponse]) error {
		if response.Payload.LockActionStatus == messages.LockOpenedLockStatus ||
			response.Payload.LockActionStatus == messages.ErrorLockAlreadyOpenLockStatus &&
				len(response.Payload.ChannelIds) > 0 {
			return nil
		}

		return errors.New("Command failed with status: " + string(response.Payload.LockActionStatus))
	}
	ctx, cancel := context.WithTimeout(parent, 7*time.Second)
	defer cancel()

	return mqttRequestSync[messages.Response[messages.LockResponse]](s, device, ctx, responseValidator, func() (mqtt.Token, error) { return s.Auto(device, recloseDelay, channels) })
}

func (s *DeviceService) AllowKeyAccess(device *entities.Device, transactionId int, auth messages.Auth) mqtt.Token {
	auth.AuthStatus = messages.SuccessOnlineStatus
	return s.keyAuthorization(device, transactionId, auth)
}
func (s *DeviceService) AllowKeyAccessSync(device *entities.Device, transactionId int, auth messages.Auth) error {
	token := s.AllowKeyAccess(device, transactionId, auth)
	token.Wait()
	return token.Error()
}

func (s *DeviceService) DenyKeyAccessSync(device *entities.Device, transactionId int, auth messages.Auth) error {
	token := s.DenyKeyAccess(device, transactionId, auth)
	token.Wait()
	return token.Error()
}
func (s *DeviceService) DenyKeyAccess(device *entities.Device, transactionId int, auth messages.Auth) mqtt.Token {
	auth.AuthStatus = messages.FailedOnlineStatus
	return s.keyAuthorization(device, transactionId, auth)
}

func (s *DeviceService) keyAuthorization(device *entities.Device, transactionId int, auth messages.Auth) mqtt.Token {
	data, _ := json.Marshal(messages.NewAuthEvent(transactionId, auth))

	return s.pool.GetClient(services.GetClientOptions(device.Gateway.Broker).ClientID).
		Publish(device.GetCommandsTopic(), 0, false, data)
}

func (s *DeviceService) EnqueueCommand(device *entities.Device, name string, payload map[string]any) (int, error) {
	return 0, nil
}

func (s *DeviceService) Locate(device *entities.Device, transactionId int) mqtt.Token {
	message, _ := json.MarshalNoEscape(messages.NewLocateRequest(transactionId))
	return s.pool.GetClient(services.GetClientOptions(device.Gateway.Broker).ClientID).
		Publish(device.GetCommandsTopic(), 0, false, message)
}
func (s *DeviceService) LocateSync(ctx context.Context, device *entities.Device, transactionId int) error {
	token := s.Locate(device, transactionId)

	select {
	case <-ctx.Done():
		return fiber.ErrGatewayTimeout
	case <-token.Done():
		return token.Error()
	}
}

func (s *DeviceService) GetFirmware(device *entities.Device) mqtt.Token {
	message, _ := json.MarshalNoEscape(messages.NewFirmwareVersionRequest(0))

	return s.pool.GetClient(services.GetClientOptions(device.Gateway.Broker).ClientID).
		Publish(device.GetCommandsTopic(), 0, false, message)
}
func (s *DeviceService) GetFirmwareSync(parent context.Context, device *entities.Device) (string, error) {
	//responseValidator := func(response messages.Response[messages.LockResponse]) error {
	//}
	//ctx, cancel := context.WithTimeout(parent, 7*time.Second)
	//defer cancel()

	return "", nil
}

func (s *DeviceService) ReadConfig() mqtt.Token {
	return nil
}
func (s *DeviceService) ReadConfigSync(ctx context.Context, device *entities.Device) error {
	return nil
}

func (s *DeviceService) ClearQueueForDevice(ctx context.Context, device *entities.Device) {
}

func mqttRequestSync[T any](s *DeviceService, device *entities.Device, ctx context.Context, validate func(T) error, action func() (mqtt.Token, error)) error {
	errorChan := make(chan error)

	handler := application.HandlerFunc(func(client mqtt.Client, message mqtt.Message) {
		if message.Topic() != device.GetEventsTopic() {
			return
		}

		var response T
		if err := json.UnmarshalNoEscape(message.Payload(), &response); err != nil {
			errorChan <- err
			return
		}

		if err := validate(response); err != nil {
			errorChan <- err
			return
		}

		errorChan <- nil
	})

	s.pool.Register(handler)
	defer s.pool.Unregister(handler)
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
		message, _ := json.MarshalNoEscape(messages.NewLockOfflineResponse())
		s.pool.GetClient(services.GetClientOptions(device.Gateway.Broker).ClientID).Publish(device.GetEventsTopic(), 2, false, message)

		return fiber.ErrGatewayTimeout
	case err = <-errorChan:
		return err
	}
}
