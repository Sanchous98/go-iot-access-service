package listeners

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/application/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/messages"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"log"
	"regexp"
	"strings"
)

type VerifyOnlineHandler struct {
	regexp            *regexp.Regexp
	deviceRepository  application.DeviceRepository  `inject:""`
	gatewayRepository application.GatewayRepository `inject:""`
	coreApiBaseUrl    string                        `env:"CORE_API_SERVICE_URL"`
	coreApiKey        string                        `env:"CORE_API_SERVICE_ACCESS_TOKEN"`
	deviceService     *services.DeviceService       `inject:""`
	client            *fiber.Client
}

func (h *VerifyOnlineHandler) Constructor() {
	h.regexp = regexp.MustCompile(`^\$foursuites/gw/(.+)/dev/(.+)/events$`)
	h.client = fiber.AcquireClient()
}

func (h *VerifyOnlineHandler) Handle(_ mqtt.Client, message mqtt.Message) {
	var p messages.Response[messages.Auth]
	_ = json.UnmarshalNoEscape(message.Payload(), &p)

	p.Payload.HashKey = strings.TrimPrefix(p.Payload.HashKey, "0x")

	res := h.regexp.FindStringSubmatch(message.Topic())
	gatewayMacId := res[1]
	deviceMacId := res[2]
	device := h.deviceRepository.FindByMacId(deviceMacId)

	if device == nil {
		return
	}

	device.Gateway = h.gatewayRepository.FindByIeee(gatewayMacId)
	response := h.authorizationRequest(deviceMacId, gatewayMacId, p.Payload.HashKey, p.Payload.AuthType)

	if response == nil || response.AccessibleChannels == nil || len(response.AccessibleChannels) == 0 {
		err := h.deviceService.DenyKeyAccessSync(device, 0, messages.Auth{
			HashKey:  p.Payload.HashKey,
			AuthType: p.Payload.AuthType,
		})

		if err != nil {
			mqtt.ERROR.Println(err)
		}

		return
	}

	err := h.deviceService.AllowKeyAccessSync(device, 0, messages.Auth{
		HashKey:    p.Payload.HashKey,
		AuthType:   p.Payload.AuthType,
		ChannelIds: response.AccessibleChannels,
	})

	if err != nil {
		mqtt.ERROR.Println(err)
	}
}

func (h *VerifyOnlineHandler) CanHandle(_ mqtt.Client, message mqtt.Message) bool {
	var p messages.Response[messages.Auth]
	err := json.UnmarshalNoEscape(message.Payload(), &p)
	return err == nil && h.regexp.MatchString(message.Topic()) && p.Payload.AuthStatus == messages.VerifyOnlineStatus
}

type authorizationResponsePayload struct {
	AccessibleChannels []int `json:"accessibleChannels"`
}

func (h *VerifyOnlineHandler) authorizationRequest(deviceMacId, gatewayMacId, hashKey string, authTypes messages.AuthType) *authorizationResponsePayload {
	var authTypeParam int

	if authTypes != messages.NfcType {
		authTypeParam = 1
	}

	arg := fiber.AcquireArgs()

	arg.Set("device_mac_id", deviceMacId)
	arg.Set("gateway_mac_id", gatewayMacId)
	arg.Set("qr_crc_hash", hashKey)
	arg.Set("auth_type", utils.ToString(authTypeParam))

	agent := h.client.Post(h.coreApiBaseUrl+"/qr-codes/qr-scan").
		Add(fiber.HeaderAccept, fiber.MIMEApplicationJSON).
		Add(fiber.HeaderAuthorization, "Bearer "+h.coreApiKey).
		Form(arg)

	var responseData struct {
		Data *authorizationResponsePayload `json:"data"`
	}

	code, body, errors := agent.Struct(&responseData)
	fiber.ReleaseArgs(arg)

	if len(errors) != 0 {
		log.Println(errors)
		return nil
	}

	if code >= 400 {
		log.Println(utils.UnsafeString(body))
		return nil
	}

	return responseData.Data
}
