package controllers

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"github.com/gofiber/fiber/v2"
	"log"
)

// NetworkOpen => POST /gateways/:gatewayId/network/open
func NetworkOpen(repository repositories.GatewayRepository, service services.GatewayService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		gateway := repository.FindByIeee(ctx.Params("gatewayId"))

		if err := service.OpenNetwork(ctx.UserContext(), gateway); err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(200)
	}
}

// NetworkClose => POST /gateways/:gatewayId/network/close
func NetworkClose(repository repositories.GatewayRepository, service services.GatewayService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		gateway := repository.FindByIeee(ctx.Params("gatewayId"))

		if err := service.CloseNetwork(ctx.UserContext(), gateway); err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(200)
	}
}

// NetworkGetState => POST /gateways/:gatewayId/network/get-state
// TODO: Why POST?!
func NetworkGetState(repository repositories.GatewayRepository, service services.GatewayService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		gateway := repository.FindByIeee(ctx.Params("gatewayId"))

		if err := service.GetNetworkState(ctx.UserContext(), gateway); err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(200)
	}
}

// DeleteDeviceFromGateway => DELETE /gateways/:gatewayId/devices/:deviceId
func DeleteDeviceFromGateway(repository repositories.GatewayRepository, service services.GatewayService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		gateway := repository.FindByIeee(ctx.Params("gatewayId"))

		if err := service.RemoveDeviceSync(ctx.UserContext(), gateway, ctx.Params("deviceId")); err != nil {
			log.Println(err)
			return fiber.ErrInternalServerError
		}

		return ctx.SendStatus(200)
	}
}
