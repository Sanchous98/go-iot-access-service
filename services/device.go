package services

import (
	"bitbucket.org/4suites/iot-service-golang/messages"
	"bitbucket.org/4suites/iot-service-golang/models"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

type DeviceService struct {
	aggregator *HandlerAggregator[*models.Broker] `inject:""`
}

func (s *DeviceService) Open(device *models.Device, channels []int) mqtt.Token {
	message, _ := json.MarshalNoEscape(messages.NewLockOpenEvent(0, channels))

	return s.aggregator.GetClient(device.GetGateway().GetBroker()).Publish(device.GetCommandsTopic(), 2, false, message)
}
func (s *DeviceService) OpenSync(device *models.Device, channels []int) error {
	responseHandler := func(response messages.EventResponse[messages.LockResponse]) bool {
		return response.Event.Payload.LockActionStatus == messages.LockOpenedLockStatus ||
			response.Event.Payload.LockActionStatus == messages.ErrorLockAlreadyOpenLockStatus &&
				len(response.Event.Payload.ChannelIds) > 0
	}

	return s.waitForResponse(device, func() mqtt.Token { return s.Open(device, channels) }, responseHandler)
}

func (s *DeviceService) Close(device *models.Device) mqtt.Token {
	message, _ := json.MarshalNoEscape(messages.NewLockCloseEvent(0))

	return s.aggregator.GetClient(device.GetGateway().GetBroker()).Publish(device.GetCommandsTopic(), 2, false, message)
}
func (s *DeviceService) CloseSync(device *models.Device) error {
	responseHandler := func(response messages.EventResponse[messages.LockResponse]) bool {
		return response.Event.Payload.LockActionStatus == messages.LockClosedLockStatus ||
			response.Event.Payload.LockActionStatus == messages.ErrorLockAlreadyClosedLockStatus &&
				len(response.Event.Payload.ChannelIds) > 0
	}

	return s.waitForResponse(device, func() mqtt.Token { return s.Close(device) }, responseHandler)
}

func (s *DeviceService) Auto(device *models.Device, recloseDelay byte, channels []int) mqtt.Token {
	message, _ := json.Marshal(messages.NewLockAutoEvent(0, recloseDelay, channels))

	return s.aggregator.GetClient(device.GetGateway().GetBroker()).Publish(device.GetCommandsTopic(), 2, false, message)
}
func (s *DeviceService) AutoSync(device *models.Device, recloseDelay byte, channels []int) error {
	responseHandler := func(response messages.EventResponse[messages.LockResponse]) bool {
		return response.Event.Payload.LockActionStatus == messages.LockOpenedLockStatus ||
			response.Event.Payload.LockActionStatus == messages.ErrorLockAlreadyOpenLockStatus &&
				len(response.Event.Payload.ChannelIds) > 0
	}

	return s.waitForResponse(device, func() mqtt.Token { return s.Auto(device, recloseDelay, channels) }, responseHandler)
}

func (s *DeviceService) AllowKeyAccess(device *models.Device, transactionId int, payload map[string]any) <-chan mqtt.Token {
	return s.KeyAuthorization(device, transactionId, payload["hashKey"].(string), payload["authType"].(messages.AuthType), messages.SuccessOnlineStatus, payload)
}
func (s *DeviceService) AllowKeyAccessSync(device *models.Device, transactionId int, payload map[string]any) error {
	select {
	case token := <-s.AllowKeyAccess(device, transactionId, payload):
		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	return nil
}

func (s *DeviceService) DenyKeyAccessSync(device *models.Device, transactionId int, payload map[string]any) error {
	select {
	case token := <-s.DenyKeyAccess(device, transactionId, payload):
		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	return nil
}
func (s *DeviceService) DenyKeyAccess(device *models.Device, transactionId int, payload map[string]any) <-chan mqtt.Token {
	return s.KeyAuthorization(device, transactionId, payload["hashKey"].(string), payload["authType"].(messages.AuthType), messages.FailedOnlineStatus, payload)
}

func (s *DeviceService) KeyAuthorization(device *models.Device, transactionId int, hashKey string, authType messages.AuthType, authStatus messages.AuthStatus, params map[string]any) <-chan mqtt.Token {
	auth := messages.Auth{
		HashKey:    hashKey,
		AuthType:   authType,
		AuthStatus: authStatus,
	}

	if channelIds, ok := params["channelIds"]; ok {
		auth.ChannelIds = channelIds.([]int)
	}

	data, _ := json.Marshal(messages.NewAuthEvent(transactionId, auth))
	log.Println(string(data))
	return s.aggregator.Publish(device.GetGateway().GetBroker(), data, 0)
}

func (s *DeviceService) waitForResponse(device *models.Device, action func() mqtt.Token, lockOpened func(response messages.EventResponse[messages.LockResponse]) bool) error {
	errorChan := make(chan error)
	var handler HandlerFunc

	handler = func(client mqtt.Client, message mqtt.Message) {
		if message.Topic() != device.GetEventsTopic() {
			return
		}

		defer close(errorChan)

		var response messages.EventResponse[messages.LockResponse]
		err := json.UnmarshalNoEscape(message.Payload(), &response)

		if err != nil {
			errorChan <- err
			return
		}

		if response.Event.EventType == messages.LockActionResponse &&
			response.TransactionId == 0 &&
			response.Event.Payload.LockActionStatus != messages.ExtRelayStateLockStatus {
			s.aggregator.Unregister(handler)
		}

		if !lockOpened(response) {
			errorChan <- errors.New("Command failed with status: " + string(response.Event.Payload.LockActionStatus))
		}
	}

	s.aggregator.Register(handler)
	defer s.aggregator.Unregister(handler)
	token := action()

	if !token.WaitTimeout(7 * time.Second) {
		message, _ := json.MarshalNoEscape(messages.NewLockOfflineResponse())
		s.aggregator.GetClient(device.GetGateway().GetBroker()).Publish(device.GetEventsTopic(), 2, false, message)

		return fiber.ErrGatewayTimeout
	}

	select {
	case err := <-errorChan:
		return err
	default:
		return nil
	}
}
