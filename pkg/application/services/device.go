package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/messages"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/unsafe"
	"context"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"time"
)

type DeviceService struct {
	pool application.ClientPool `inject:""`
	log  logger.Logger          `inject:""`
}

func (s *DeviceService) Open(device *entities.Device, channels []int) (mqtt.Token, error) {
	message := unsafe.Suppress(json.MarshalNoEscape(messages.NewLockOpenEvent(0, channels)))

	client := s.pool.GetClient(application.GetClientOptions(s.log, device.Gateway.Broker).ClientID)

	if client == nil {
		s.log.Errorf("Client %s does not exist\n", application.GetClientOptions(s.log, device.Gateway.Broker).ClientID)
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

	return mqttRequestSync[messages.Response[messages.LockResponse]](s, device, ctx, responseValidator, func() (mqtt.Token, error) { return s.Open(device, channels) }, s.log)
}

func (s *DeviceService) Close(device *entities.Device) (mqtt.Token, error) {
	message := unsafe.Suppress(json.MarshalNoEscape(messages.NewLockCloseEvent(0)))

	client := s.pool.GetClient(application.GetClientOptions(s.log, device.Gateway.Broker).ClientID)

	if client == nil {
		s.log.Errorf("Client %s does not exist\n", application.GetClientOptions(s.log, device.Gateway.Broker).ClientID)
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

	return mqttRequestSync[messages.Response[messages.LockResponse]](s, device, ctx, responseValidator, func() (mqtt.Token, error) { return s.Close(device) }, s.log)
}

func (s *DeviceService) Auto(device *entities.Device, recloseDelay byte, channels []int) (mqtt.Token, error) {
	message := unsafe.Suppress(json.MarshalNoEscape(messages.NewLockAutoEvent(0, recloseDelay, channels)))

	client := s.pool.GetClient(application.GetClientOptions(s.log, device.Gateway.Broker).ClientID)

	if client == nil {
		s.log.Errorf("Client %s does not exist\n", application.GetClientOptions(s.log, device.Gateway.Broker).ClientID)
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

	return mqttRequestSync[messages.Response[messages.LockResponse]](s, device, ctx, responseValidator, func() (mqtt.Token, error) { return s.Auto(device, recloseDelay, channels) }, s.log)
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
	data := unsafe.Suppress(json.MarshalNoEscape(messages.NewAuthEvent(transactionId, auth)))

	return s.pool.GetClient(application.GetClientOptions(s.log, device.Gateway.Broker).ClientID).
		Publish(device.GetCommandsTopic(), 0, false, data)
}

func (s *DeviceService) EnqueueCommand(device *entities.Device, name string, payload map[string]any) (int, error) {
	return 0, nil
}

func (s *DeviceService) Locate(device *entities.Device, transactionId int) mqtt.Token {
	message := unsafe.Suppress(json.MarshalNoEscape(messages.NewLocateRequest(transactionId)))
	return s.pool.GetClient(application.GetClientOptions(s.log, device.Gateway.Broker).ClientID).
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
	message := unsafe.Suppress(json.MarshalNoEscape(messages.NewFirmwareVersionRequest(0)))

	return s.pool.GetClient(application.GetClientOptions(s.log, device.Gateway.Broker).ClientID).
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
