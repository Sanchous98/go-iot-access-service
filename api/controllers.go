package api

import (
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"bitbucket.org/4suites/iot-service-golang/services"
	"github.com/Sanchous98/go-di"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
	"sync"
)

var cache sync.Map

// TODO: Remove after demo
func convertCoreApiIdToRegistryMacId(deviceCoreId int) string {
	if item, hit := cache.Load(deviceCoreId); hit {
		return item.(string)
	}

	db, err := sqlx.Open("mysql", di.Application().GetParam("DATABASE_DSN"))
	if err != nil {
		log.Println(err)
		return ""
	}

	defer db.Close()
	var macId []string
	err = db.Select(&macId, "SELECT mac_id FROM devices WHERE id = ?", deviceCoreId)
	if err != nil {
		log.Println(err)
		return ""
	}

	if len(macId) == 0 {
		return ""
	}

	return macId[0]
}

// Open POST: /devices/:deviceId/open
func Open(service services.DeviceService, repository *repositories.DeviceRepository) fiber.Handler {
	type responseShape struct {
		ChannelsIds []int `json:"channelIds"`
	}

	return func(ctx *fiber.Ctx) error {
		var deviceId string
		coreDeviceId, err := strconv.Atoi(ctx.Params("deviceId", ""))

		if err == nil {
			deviceId = convertCoreApiIdToRegistryMacId(coreDeviceId)
		} else {
			deviceId = ctx.Params("deviceId", "")
		}

		device := repository.FindByMacId(deviceId)

		if device == nil {
			return fiber.ErrNotFound
		}

		var body responseShape

		err = ctx.BodyParser(&body)
		if err != nil {
			log.Println(err)
			return fiber.ErrUnprocessableEntity
		}

		err = service.OpenSync(device, body.ChannelsIds)
		if err != nil {
			log.Println(err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		ctx.Status(200)
		return ctx.JSON(map[string]int{"status": 200})
	}
}

// Close POST: /devices/:deviceId/close
func Close(service services.DeviceService, repository *repositories.DeviceRepository) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var deviceId string
		coreDeviceId, err := strconv.Atoi(ctx.Params("deviceId", ""))

		if err == nil {
			deviceId = convertCoreApiIdToRegistryMacId(coreDeviceId)
		} else {
			log.Println(err)
			deviceId = ctx.Params("deviceId", "")
		}

		device := repository.FindByMacId(deviceId)

		if device == nil {
			return fiber.ErrNotFound
		}

		err = service.CloseSync(device)
		if err != nil {
			log.Println(err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		ctx.Status(200)
		return ctx.JSON(map[string]int{"status": 200})
	}
}

// Auto POST: /devices/:deviceId/auto
func Auto(service services.DeviceService, repository *repositories.DeviceRepository) fiber.Handler {
	type responseShape struct {
		RecloseDelay uint8 `json:"recloseDelay"`
		ChannelsIds  []int `json:"channelIds"`
	}

	return func(ctx *fiber.Ctx) error {
		var deviceId string
		coreDeviceId, err := strconv.Atoi(ctx.Params("deviceId", ""))

		if err == nil {
			deviceId = convertCoreApiIdToRegistryMacId(coreDeviceId)
		} else {
			log.Println(err)
			deviceId = ctx.Params("deviceId", "")
		}

		device := repository.FindByMacId(deviceId)

		if device == nil {
			return fiber.ErrNotFound
		}

		var body responseShape

		err = ctx.BodyParser(&body)
		if err != nil {
			log.Println(err)
			return fiber.ErrUnprocessableEntity
		}

		err = service.AutoSync(device, body.RecloseDelay, body.ChannelsIds)
		if err != nil {
			log.Println(err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		ctx.Status(200)
		return ctx.JSON(map[string]int{"status": 200})
	}
}
