package controllers

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/repositories"
	"github.com/gofiber/fiber/v2"
)

type requestShape struct {
	CommandName string         `json:"name"`
	Payload     map[string]any `json:"payload"`
}

// Commands => POST /devices/:deviceId/commands
func Commands(repository repositories.DeviceRepository, service services.DeviceService, log logger.Logger) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			log.Debugf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		var body requestShape

		if err := ctx.BodyParser(&body); err != nil {
			ctx.Status(fiber.StatusBadRequest)
			return err
		}

		commandId, err := service.EnqueueCommand(device, body.CommandName, body.Payload)

		if err != nil {
			log.Errorln(err)
			return fiber.ErrInternalServerError
		}

		return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
			"commandId": commandId,
			"name":      body.CommandName,
			"payload":   body.Payload,
		})
	}
}
