package api

import (
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"bitbucket.org/4suites/iot-service-golang/services"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"log"
)

type requestShape struct {
	RecloseDelay uint8 `json:"recloseDelay,omitempty"`
	ChannelsIds  []int `json:"channelIds,omitempty"`
}

// Action => POST /devices/:deviceId/:action
func Action(service services.DeviceService, repository *repositories.DeviceRepository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			return fiber.ErrNotFound
		}

		var body requestShape

		err := ctx.BodyParser(&body)
		if err != nil {
			log.Println(err)
			return fiber.ErrUnprocessableEntity
		}

		switch ctx.Params("action") {
		case "open":
			err = service.OpenSync(device, body.ChannelsIds)
		case "close":
			err = service.CloseSync(device)
		case "auto":
			err = service.AutoSync(device, body.RecloseDelay, body.ChannelsIds)
		}

		if err != nil {
			log.Println(err)

			if _, ok := err.(*fiber.Error); ok {
				return err
			}

			return fiber.ErrInternalServerError
		}

		return ctx.Status(200).JSON(fiber.Map{"status": 200})
	}
}
