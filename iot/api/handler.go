package api

import (
	"bitbucket.org/4suites/iot-service-golang/iot/api/controllers"
	"bitbucket.org/4suites/iot-service-golang/pkg/application/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
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
	log               logger.Logger                  `inject:""`
}

func (h *IotApiHandler) RegisterRoutes(app *fiber.App) {
	app.Delete(
		"/gateways/:gatewayId/devices/:deviceId",
		controllers.DeleteDeviceFromGateway(h.gatewayRepository, h.gatewayService, h.log),
	)

	{
		gateways := app.Group("/gateways/:gatewayId/network")
		gateways.Post(
			"/open",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.NetworkOpen(h.gatewayRepository, h.gatewayService, h.log),
		)
		gateways.Post(
			"/close",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.NetworkClose(h.gatewayRepository, h.gatewayService, h.log),
		)
		gateways.Get(
			"/get-state",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.NetworkGetState(h.gatewayRepository, h.gatewayService, h.log),
		)
		gateways.Post(
			"/get-state",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.NetworkGetState(h.gatewayRepository, h.gatewayService, h.log),
		)
	}

	{
		devices := app.Group("/devices/:deviceId")
		devices.Post(
			"/locate",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.Locate(h.deviceRepository, h.deviceService, h.log),
		)

		devices.Get(
			"/:gatewayId/firmware-version",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.FirmwareVersion(h.deviceRepository, h.deviceService, h.log),
		)

		devices.Get(
			"/:gatewayId/config",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.Config(h.deviceRepository, h.deviceService, h.log),
		)

		devices.Post(
			"/keys",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.DeviceCreateOfflineKey(h.deviceRepository, h.deviceService, h.log),
		)

		devices.Put(
			"/keys/:hashKey",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.DeviceUpdateOfflineKey(h.deviceRepository, h.deviceService, h.log),
		)

		devices.Delete(
			"/keys",
			middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
			controllers.DeviceDeleteOfflineKey(h.deviceRepository, h.deviceService, h.log),
		)
	}

	app.Post(
		"/commands",
		middleware.ConvertCoreDeviceIdToRegistryMac(h.db, h.log),
		controllers.Commands(h.deviceRepository, h.deviceService, h.log),
	)
}
