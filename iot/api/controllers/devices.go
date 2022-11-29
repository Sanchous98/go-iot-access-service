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
