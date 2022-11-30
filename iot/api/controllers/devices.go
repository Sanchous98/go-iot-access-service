package controllers

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"github.com/gofiber/fiber/v2"
	"log"
)

// Locate => POST /devices/:deviceId/locate
func Locate(repository repositories.DeviceRepository, service services.DeviceService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			log.Printf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		if err := service.LocateSync(device, 0); err != nil {
			log.Println(err)

			return fiber.ErrInternalServerError
		}

		return ctx.Status(fiber.StatusOK).SendString("Locate command has been sent!")
	}
}

// FirmwareVersion => GET /devices/:deviceMacId/:gatewayMacId/firmware-version
func FirmwareVersion(repository repositories.DeviceRepository, service services.DeviceService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

	}
}

// Config => GET /devices/:deviceMacId/:gatewayMacId/config
func Config(repository repositories.DeviceRepository, service services.DeviceService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

	}
}

// DeviceCreateOfflineKey => POST /devices/:deviceId/keys
func DeviceCreateOfflineKey(repository repositories.DeviceRepository, service services.DeviceService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

	}
}

// DeviceUpdateOfflineKey => PUT /devices/:deviceId/keys
func DeviceUpdateOfflineKey(repository repositories.DeviceRepository, service services.DeviceService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

	}
}

// DeviceDeleteOfflineKey => DELETE /devices/:deviceId/keys
func DeviceDeleteOfflineKey(repository repositories.DeviceRepository, service services.DeviceService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

	}
}
