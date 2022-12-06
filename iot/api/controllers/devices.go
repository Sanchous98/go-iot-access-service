package controllers

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/repositories"
	"github.com/gofiber/fiber/v2"
)

// Locate => POST /devices/:deviceId/locate
func Locate(repository repositories.DeviceRepository, service services.DeviceService, log logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			log.Debugf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		if err := service.LocateSync(ctx.UserContext(), device, 0); err != nil {
			log.Errorln(err)

			return fiber.ErrInternalServerError
		}

		return ctx.Status(fiber.StatusOK).SendString("Locate command has been sent!")
	}
}

// FirmwareVersion => GET /devices/:deviceId/:gatewayId/firmware-version
func FirmwareVersion(repository repositories.DeviceRepository, service services.DeviceService, log logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacIdAndGatewayIeee(ctx.Params("deviceId"), ctx.Params("gatewayId"))

		if device == nil {
			log.Debugf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		_, err := service.GetFirmwareSync(ctx.UserContext(), device)

		if err != nil {
			log.Errorln(err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}

// Config => GET /devices/:deviceId/:gatewayId/config
func Config(repository repositories.DeviceRepository, service services.DeviceService, log logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacIdAndGatewayIeee(ctx.Params("deviceId"), ctx.Params("gatewayId"))

		if device == nil {
			log.Debugf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		if err := service.ReadConfigSync(ctx.UserContext(), device); err != nil {
			log.Errorln(err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}

// DeviceCreateOfflineKey => POST /devices/:deviceId/keys
func DeviceCreateOfflineKey(repository repositories.DeviceRepository, service services.DeviceService, log logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			log.Debugf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		var data map[string]any
		_ = ctx.BodyParser(&data)

		commandId, err := service.EnqueueCommand(device, "createKey", data)

		if err != nil {
			log.Errorln(err)
			return fiber.ErrInternalServerError
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"commandId": commandId,
			"name":      "createKey",
			"payload":   data,
		})
	}
}

// DeviceUpdateOfflineKey => PUT /devices/:deviceId/keys/:hashKey
func DeviceUpdateOfflineKey(repository repositories.DeviceRepository, service services.DeviceService, log logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			log.Debugf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		var data map[string]any
		_ = ctx.BodyParser(&data)

		data["hashKey"] = ctx.Params("hashKey")
		commandId, err := service.EnqueueCommand(device, "updateKey", data)

		if err != nil {
			log.Errorln(err)
			return fiber.ErrInternalServerError
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"commandId": commandId,
			"name":      "updateKey",
			"payload":   data,
		})
	}
}

// DeviceDeleteOfflineKey => DELETE /devices/:deviceId/keys
func DeviceDeleteOfflineKey(repository repositories.DeviceRepository, service services.DeviceService, log logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			log.Debugf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		var data map[string]any
		_ = ctx.BodyParser(&data)

		if data["hashKey"] == "all" {
			service.ClearQueueForDevice(ctx.UserContext(), device)
		}

		commandId, err := service.EnqueueCommand(device, "deleteKey", data)

		if err != nil {
			log.Errorln(err)
			return fiber.ErrInternalServerError
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"commandId": commandId,
			"name":      "deleteKey",
			"payload":   data,
		})
	}
}
