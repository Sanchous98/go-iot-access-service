package api

import (
	"bitbucket.org/4suites/iot-service-golang/iot/api/controllers"
	"bitbucket.org/4suites/iot-service-golang/pkg/http/middleware"
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type IotApiHandler struct {
	deviceRepository  repositories.DeviceRepository  `inject:""`
	deviceService     services.DeviceService         `inject:""`
	gatewayRepository repositories.GatewayRepository `inject:""`
	gatewayService    services.GatewayService        `inject:""`
	db                *gorm.DB                       `inject:""`
}

func (h *IotApiHandler) RegisterRoutes(app *fiber.App) {
	app.Delete(
		"/gateways/:gatewayId/devices/:deviceId",
		controllers.DeleteDeviceFromGateway(h.gatewayRepository, h.gatewayService),
	)

	{
		gateways := app.Group("/gateways/:gatewayId/network")
		gateways.Post(
			"/open",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.NetworkOpen(h.gatewayRepository, h.gatewayService),
		)
		gateways.Post(
			"/close",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.NetworkClose(h.gatewayRepository, h.gatewayService),
		)
		gateways.Get(
			"/get-state",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.NetworkGetState(h.gatewayRepository, h.gatewayService),
		)
		gateways.Post(
			"/get-state",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.NetworkGetState(h.gatewayRepository, h.gatewayService),
		)
	}

	{
		devices := app.Group("/devices/:deviceId")
		devices.Post(
			"/locate",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.Locate(h.deviceRepository, h.deviceService),
		)

		app.Post(
			"/commands",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.Commands(h.deviceRepository, h.deviceService),
		)
	}
}
