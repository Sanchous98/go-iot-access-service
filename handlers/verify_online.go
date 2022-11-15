package handlers

import (
	"bitbucket.org/4suites/iot-service-golang/messages"
	"bitbucket.org/4suites/iot-service-golang/models"
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"bitbucket.org/4suites/iot-service-golang/services"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"log"
	"regexp"
	"strings"
)

type authorizationResponsePayload struct {
	AccessibleChannels []int `json:"accessibleChannels"`
}

type VerifyOnlineHandler struct {
	regexp            *regexp.Regexp
	deviceRepository  repositories.DeviceRepository  `inject:""`
	gatewayRepository repositories.GatewayRepository `inject:""`
	coreApiBaseUrl    string                         `env:"CORE_API_SERVICE_URL"`
	coreApiKey        string                         `env:"CORE_API_SERVICE_ACCESS_TOKEN"`
	deviceService     *services.DeviceService        `inject:""`
}

func (h *VerifyOnlineHandler) authorizationRequest(deviceMacId, gatewayMacId, hashKey string, authTypes messages.AuthType) *authorizationResponsePayload {
	var authTypeParam int

	if authTypes != messages.NfcType {
		authTypeParam = 1
	}

	body, _ := json.Marshal(map[string]any{
		"device_mac_id":  deviceMacId,
		"gateway_mac_id": gatewayMacId,
		"qr_crc_hash":    hashKey,
		"auth_type":      authTypeParam,
	})

	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	agent := client.Post(h.coreApiBaseUrl+"/qr-codes/qr-scan").
		Add(fiber.HeaderAccept, fiber.MIMEApplicationJSON).
		Add(fiber.HeaderAuthorization, "Bearer "+h.coreApiKey).
		Body(body)
	defer fiber.ReleaseAgent(agent)

	code, body, errors := agent.Bytes()

	if len(errors) != 0 || code >= 400 {
		log.Println(errors)
		return nil
	}

	responseData := struct {
		Data *authorizationResponsePayload `json:"data"`
	}{}

	_ = json.Unmarshal(body, &responseData)

	return responseData.Data
}

func (h *VerifyOnlineHandler) Constructor() {
	h.regexp = regexp.MustCompile(`^\$foursuites/gw/(.+)/dev/(.+)/events$`)
}

func (h *VerifyOnlineHandler) Handle(_ mqtt.Client, message mqtt.Message) {
	var p messages.EventRequest[messages.Auth]
	_ = json.Unmarshal(message.Payload(), &p)

	p.Event.Payload.HashKey = strings.TrimPrefix(p.Event.Payload.HashKey, "0x")

	var gatewayMacId, deviceMacId string
	_, _ = fmt.Sscanf(message.Topic(), "$foursuites/gw/%s/dev/%s/events", &gatewayMacId, &deviceMacId)
	device := h.deviceRepository.FindByMacId(deviceMacId)

	if device == nil {
		return
	}

	gateway := h.gatewayRepository.FindByMacId(gatewayMacId)
	device.GatewayResolver = func() *models.Gateway { return gateway }
	response := h.authorizationRequest(deviceMacId, gatewayMacId, p.Event.Payload.HashKey, p.Event.Payload.AuthType)

	if len(response.AccessibleChannels) == 0 {
		err := h.deviceService.DenyKeyAccessSync(device, 0, map[string]any{
			"hashKey":  p.Event.Payload.HashKey,
			"authType": p.Event.Payload.AuthType,
		})

		if err != nil {
			mqtt.ERROR.Println(err)
		}

		return
	}

	err := h.deviceService.AllowKeyAccessSync(device, 0, map[string]any{
		"hashKey":    p.Event.Payload.HashKey,
		"authType":   p.Event.Payload.AuthType,
		"channelIds": response.AccessibleChannels,
	})

	if err != nil {
		mqtt.ERROR.Println(err)
	}
}

func (h *VerifyOnlineHandler) CanHandle(_ mqtt.Client, message mqtt.Message) bool {
	var p messages.EventRequest[messages.Auth]
	err := json.Unmarshal(message.Payload(), &p)
	return err == nil && h.regexp.MatchString(message.Topic()) && p.Event.Payload.AuthStatus == "verifyOnline"
}
