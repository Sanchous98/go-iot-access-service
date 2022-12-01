package api

import (
	"bitbucket.org/4suites/iot-service-golang/iot/api/controllers"
	"bitbucket.org/4suites/iot-service-golang/pkg/application/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/http/middleware"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/repositories"
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

		devices.Get(
			"/:gatewayId/firmware-version",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.FirmwareVersion(h.deviceRepository, h.deviceService),
		)

		devices.Get(
			"/:gatewayId/config",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.Config(h.deviceRepository, h.deviceService),
		)

		devices.Post(
			"/keys",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.DeviceCreateOfflineKey(h.deviceRepository, h.deviceService),
		)

		devices.Put(
			"/keys/:hashKey",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.DeviceUpdateOfflineKey(h.deviceRepository, h.deviceService),
		)

		devices.Delete(
			"/keys",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.DeviceDeleteOfflineKey(h.deviceRepository, h.deviceService),
		)

		app.Post(
			"/commands",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
			controllers.Commands(h.deviceRepository, h.deviceService),
		)
	}
}
