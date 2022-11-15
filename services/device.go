package services

import (
	"bitbucket.org/4suites/iot-service-golang/messages"
	"bitbucket.org/4suites/iot-service-golang/models"
	"encoding/json"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

var CommandTimeout = errors.New("CommandTimeoutError")

type DeviceService struct {
	aggregator *HandlerAggregator[*models.Broker] `inject:""`
}

func (s *DeviceService) Open(device *models.Device, channels []int) <-chan mqtt.Token {
	tokens := make(chan mqtt.Token)
	message, _ := json.Marshal(messages.NewLockOpenEvent(0, channels))

	go func() {
		tokens <- s.aggregator.GetClient(device.GetGateway().GetBroker()).Publish(device.GetCommandsTopic(), 2, false, message)

		close(tokens)
	}()

	return tokens
}
func (s *DeviceService) OpenSync(device *models.Device, channels []int) error {
	var errorChan chan error

	handler := s.sync(device, func(response messages.EventResponse[messages.LockResponse]) bool {
		return response.Event.Payload.LockActionStatus == messages.LockOpenedLockStatus ||
			response.Event.Payload.LockActionStatus == messages.ErrorLockAlreadyOpenLockStatus &&
				len(response.Event.Payload.ChannelIds) > 0
	}, errorChan)

	s.aggregator.Register(handler)
	defer s.aggregator.Unregister(handler)
	tokens := s.Open(device, channels)

	select {
	case err := <-errorChan:
		return err
	case token := <-tokens:
		if !token.WaitTimeout(7 * time.Second) {
			client := s.aggregator.GetClient(device.GetGateway().GetBroker())
			message, _ := json.Marshal(map[string]any{
				"event": map[string]any{
					"eventType": "lockOfflineResponse",
					"payload": map[string]any{
						"lockActionStatus": "openTimeoutError",
					},
				},
			})

			client.Publish(device.GetEventsTopic(), 2, false, message)

			return CommandTimeout
		}
	}

	return nil
}

func (s *DeviceService) Close(device *models.Device) <-chan mqtt.Token {
	tokens := make(chan mqtt.Token)
	message, _ := json.Marshal(messages.NewLockCloseEvent(0))

	go func() {
		tokens <- s.aggregator.GetClient(device.GetGateway().GetBroker()).Publish(device.GetCommandsTopic(), 2, false, message)

		close(tokens)
	}()

	return tokens
}
func (s *DeviceService) CloseSync(device *models.Device) error {
	var errorChan chan error
	handler := s.sync(device, func(response messages.EventResponse[messages.LockResponse]) bool {
		return response.Event.Payload.LockActionStatus == messages.LockOpenedLockStatus ||
			response.Event.Payload.LockActionStatus == messages.ErrorLockAlreadyOpenLockStatus &&
				len(response.Event.Payload.ChannelIds) > 0
	}, errorChan)

	s.aggregator.Register(handler)
	defer s.aggregator.Unregister(handler)
	tokens := s.Close(device)

	select {
	case err := <-errorChan:
		return err
	case token := <-tokens:
		if !token.WaitTimeout(7 * time.Second) {
			client := s.aggregator.GetClient(device.GetGateway().GetBroker())
			message, _ := json.Marshal(map[string]any{
				"event": map[string]any{
					"eventType": "lockOfflineResponse",
					"payload": map[string]any{
						"lockActionStatus": "openTimeoutError",
					},
				},
			})

			client.Publish(device.GetEventsTopic(), 2, false, message)

			return CommandTimeout
		}
	}

	return nil
}

func (s *DeviceService) Auto(device *models.Device, recloseDelay uint8, channels []int) <-chan mqtt.Token {
	tokens := make(chan mqtt.Token)
	message, _ := json.Marshal(messages.NewLockAutoEvent(0, recloseDelay, channels))

	go func() {
		tokens <- s.aggregator.GetClient(device.GetGateway().GetBroker()).Publish(device.GetCommandsTopic(), 2, false, message)

		close(tokens)
	}()

	return tokens
}
func (s *DeviceService) AutoSync(device *models.Device, recloseDelay uint8, channels []int) error {
	var errorChan chan error
	handler := s.sync(device, func(response messages.EventResponse[messages.LockResponse]) bool {
		return response.Event.Payload.LockActionStatus == messages.LockOpenedLockStatus ||
			response.Event.Payload.LockActionStatus == messages.ErrorLockAlreadyOpenLockStatus &&
				len(response.Event.Payload.ChannelIds) > 0
	}, errorChan)

	s.aggregator.Register(handler)
	defer s.aggregator.Unregister(handler)
	tokens := s.Auto(device, recloseDelay, channels)

	select {
	case err := <-errorChan:
		return err
	case token := <-tokens:
		if !token.WaitTimeout(7 * time.Second) {
			client := s.aggregator.GetClient(device.GetGateway().GetBroker())
			message, _ := json.Marshal(map[string]any{
				"event": map[string]any{
					"eventType": "lockOfflineResponse",
					"payload": map[string]any{
						"lockActionStatus": "openTimeoutError",
					},
				},
			})

			client.Publish(device.GetEventsTopic(), 2, false, message)

			return CommandTimeout
		}
	}

	return nil
}

func (s *DeviceService) AllowKeyAccess(device *models.Device, transactionId int, payload map[string]any) <-chan mqtt.Token {
	return s.KeyAuthorization(device, transactionId, payload["hashKey"].(string), messages.AuthType(payload["authType"].(string)), messages.SuccessOnlineStatus, payload)
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
	return s.KeyAuthorization(device, transactionId, payload["hashKey"].(string), messages.AuthType(payload["authType"].(string)), messages.FailedOnlineStatus, payload)
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

	return s.aggregator.Publish(device.GetGateway().GetBroker(), data, 0)
}

func (s *DeviceService) sync(device *models.Device, lockOpened func(messages.EventResponse[messages.LockResponse]) bool, errorChan chan<- error) (handler HandlerFunc) {
	return func(client mqtt.Client, message mqtt.Message) {
		if message.Topic() != device.GetEventsTopic() {
			return
		}

		var response messages.EventResponse[messages.LockResponse]
		err := json.Unmarshal(message.Payload(), &response)

		if err == nil &&
			response.Event.EventType == messages.LockActionResponse &&
			response.TransactionId == 0 &&
			response.Event.Payload.LockActionStatus != messages.ExtRelayStateLockStatus {
			s.aggregator.Unregister(handler)
		}

		if !lockOpened(response) {
			errorChan <- errors.New(string("Command failed with status: " + response.Event.Payload.LockActionStatus))
		}
	}
}
