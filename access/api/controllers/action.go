package controllers

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/application/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/repositories"
	services2 "bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/services"
	"github.com/eko/gocache/v3/cache"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
			err = service.OpenSync(ctx.UserContext(), device, body.ChannelsIds)
		case "close":
			err = service.CloseSync(ctx.UserContext(), device)
		case "auto":
			err = service.AutoSync(ctx.UserContext(), device, body.RecloseDelay, body.ChannelsIds)
		}

		if err != nil {
			log.Println(err)

			if _, ok := err.(*fiber.Error); ok {
				return err
			}

			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(200)
	}
}

// DisconnectBroker => POST /broker/:brokerId/disconnect
func DisconnectBroker(repository repositories.BrokerRepository, cache cache.CacheInterface[*entities.Broker], pool application.HandlerPool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, _ := uuid.Parse(ctx.Params("brokerId"))
		broker := repository.Find(id)
		pool.DeleteClient(services2.GetClientOptions(broker).ClientID)

		if err := cache.Delete(ctx.UserContext(), ctx.Params("brokerId")); err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}
