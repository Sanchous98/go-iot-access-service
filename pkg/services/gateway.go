package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/messages"
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

type GatewayService struct {
	aggregator *HandlerAggregator[*models.Gateway] `inject:""`
}

func (s *GatewayService) OpenNetwork(ctx context.Context, gateway *models.Gateway) error {
	token := s.updateNetworkState(gateway, messages.OpenState)
	select {
	case <-ctx.Done():
		return fiber.ErrGatewayTimeout
	case <-token.Done():
		return token.Error()
	}
}
func (s *GatewayService) CloseNetwork(ctx context.Context, gateway *models.Gateway) error {
	token := s.updateNetworkState(gateway, messages.CloseState)
	select {
	case <-ctx.Done():
		return fiber.ErrGatewayTimeout
	case <-token.Done():
		return token.Error()
	}
}
func (s *GatewayService) GetNetworkState(ctx context.Context, gateway *models.Gateway) error {
	message, _ := json.MarshalNoEscape(messages.NewNetworkInfoRequest(0))
	token := s.aggregator.GetClient(gateway).Publish(gateway.GetCommandTopic(), 0, false, message)

	select {
	case <-ctx.Done():
		return fiber.ErrGatewayTimeout
	case <-token.Done():
		return token.Error()
	}
}
func (s *GatewayService) RemoveDevice(gateway *models.Gateway, deviceId string) mqtt.Token {
	message, _ := json.MarshalNoEscape(messages.NewRemoveDeviceRequest(0, deviceId))
	return s.aggregator.GetClient(gateway).Publish(gateway.GetCommandTopic(), 0, false, message)
}
func (s *GatewayService) RemoveDeviceSync(ctx context.Context, gateway *models.Gateway, deviceId string) error {
	token := s.RemoveDevice(gateway, deviceId)

	select {
	case <-ctx.Done():
		return fiber.ErrGatewayTimeout
	case <-token.Done():
		return token.Error()
	}
}

func (s *GatewayService) updateNetworkState(gateway *models.Gateway, state messages.NetworkState) mqtt.Token {
	message, _ := json.MarshalNoEscape(messages.NewUpdateNetworkState(0, state))

	return s.aggregator.GetClient(gateway).Publish(gateway.GetCommandTopic(), 0, false, message)
}
