package handlers

import (
	"bitbucket.org/4suites/iot-service-golang/messages"
	"bitbucket.org/4suites/iot-service-golang/models"
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"bitbucket.org/4suites/iot-service-golang/services"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"log"
	"regexp"
	"strconv"
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
	client            *fiber.Client
}

func (h *VerifyOnlineHandler) authorizationRequest(deviceMacId, gatewayMacId, hashKey string, authTypes messages.AuthType) *authorizationResponsePayload {
	var authTypeParam int

	if authTypes != messages.NfcType {
		authTypeParam = 1
	}

	body, _ := json.MarshalNoEscape(map[string]any{
		"device_mac_id":  deviceMacId,
		"gateway_mac_id": gatewayMacId,
		"qr_crc_hash":    hashKey,
		"auth_type":      authTypeParam,
	})

	arg := fiber.AcquireArgs()
	defer fiber.ReleaseArgs(arg)

	arg.Set("device_mac_id", deviceMacId)
	arg.Set("gateway_mac_id", gatewayMacId)
	arg.Set("qr_crc_hash", hashKey)
	arg.Set("auth_type", strconv.Itoa(authTypeParam))

	agent := h.client.Post(h.coreApiBaseUrl+"/qr-codes/qr-scan").
		Add(fiber.HeaderAccept, fiber.MIMEApplicationJSON).
		Add(fiber.HeaderAuthorization, "Bearer "+h.coreApiKey).
		Form(arg)

	code, body, errors := agent.Bytes()

	if len(errors) != 0 {
		log.Println(errors)
		return nil
	}

	if code >= 400 {
		log.Println(body)
		return nil
	}

	responseData := struct {
		Data *authorizationResponsePayload `json:"data"`
	}{}

	_ = json.UnmarshalNoEscape(body, &responseData)

	return responseData.Data
}

func (h *VerifyOnlineHandler) Constructor() {
	h.regexp = regexp.MustCompile(`^\$foursuites/gw/(.+)/dev/(.+)/events$`)
	h.client = fiber.AcquireClient()
}

func (h *VerifyOnlineHandler) Handle(_ mqtt.Client, message mqtt.Message) {
	var p messages.Response[messages.Auth]
	_ = json.Unmarshal(message.Payload(), &p)
	log.Println(string(message.Payload()))

	p.Payload.HashKey = strings.TrimPrefix(p.Payload.HashKey, "0x")

	res := h.regexp.FindStringSubmatch(message.Topic())
	gatewayMacId := res[1]
	deviceMacId := res[2]
	device := h.deviceRepository.FindByMacId(deviceMacId)

	if device == nil {
		return
	}

	gateway := h.gatewayRepository.FindByMacId(gatewayMacId)
	device.GatewayResolver = func() *models.Gateway { return gateway }
	response := h.authorizationRequest(deviceMacId, gatewayMacId, p.Payload.HashKey, p.Payload.AuthType)

	if response == nil || response.AccessibleChannels == nil || len(response.AccessibleChannels) == 0 {
		err := h.deviceService.DenyKeyAccessSync(device, 0, map[string]any{
			"hashKey":  p.Payload.HashKey,
			"authType": p.Payload.AuthType,
		})

		if err != nil {
			mqtt.ERROR.Println(err)
		}

		return
	}

	err := h.deviceService.AllowKeyAccessSync(device, 0, map[string]any{
		"hashKey":    p.Payload.HashKey,
		"authType":   p.Payload.AuthType,
		"channelIds": response.AccessibleChannels,
	})

	if err != nil {
		mqtt.ERROR.Println(err)
	}
}

func (h *VerifyOnlineHandler) CanHandle(_ mqtt.Client, message mqtt.Message) bool {
	var p messages.Response[messages.Auth]
	err := json.Unmarshal(message.Payload(), &p)
	return err == nil && h.regexp.MatchString(message.Topic()) && p.Payload.AuthStatus == messages.VerifyOnlineStatus
}
