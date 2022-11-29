package api

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"github.com/gofiber/fiber/v2"
	"log"
)

type requestShape struct {
	RecloseDelay uint8 `json:"recloseDelay,omitempty"`
	ChannelsIds  []int `json:"channelIds,omitempty"`
}

// Action => POST /devices/:deviceId/:action
func Action(service services.DeviceService, repository repositories.DeviceRepository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		device := repository.FindByMacId(ctx.Params("deviceId"))

		if device == nil {
			log.Printf("Device %s not found\n", ctx.Params("deviceId"))
			return fiber.ErrNotFound
		}

		var body requestShape
		var err error

		_ = ctx.BodyParser(&body)
		//if err != nil {
		//	log.Println(err)
		//	return fiber.ErrUnprocessableEntity
		//}

		switch ctx.Params("action") {
		case "open":
			service.Open(device, body.ChannelsIds)
		case "close":
			service.Close(device)
		case "auto":
			service.Auto(device, body.RecloseDelay, body.ChannelsIds)
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
