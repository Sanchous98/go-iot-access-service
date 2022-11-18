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
var db *sqlx.DB

type responseShape struct {
	RecloseDelay uint8 `json:"recloseDelay,omitempty"`
	ChannelsIds  []int `json:"channelIds,omitempty"`
}

// TODO: Remove after demo
func convertCoreApiIdToRegistryMacId(deviceCoreId int) string {
	if item, hit := cache.Load(deviceCoreId); hit {
		return item.(string)
	}

	var err error

	if db == nil {
		db, err = sqlx.Open("mysql", di.Application().GetParam("DATABASE_DSN"))
		if err != nil {
			log.Println(err)
			return ""
		}
	}

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

// Action => POST /devices/:deviceId/:action
func Action(service services.DeviceService, repository *repositories.DeviceRepository) fiber.Handler {
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
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		ctx.Status(200)
		return ctx.JSON(map[string]int{"status": 200})
	}
}
