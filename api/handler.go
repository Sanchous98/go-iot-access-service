package api

import (
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"bitbucket.org/4suites/iot-service-golang/services"
	"database/sql"
	"errors"
	"github.com/Sanchous98/go-di"
	"github.com/goccy/go-reflect"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
	"sync"
	"unsafe"
)

type Handler struct {
	service    services.DeviceService         `inject:""`
	repository *repositories.DeviceRepository `inject:""`
	server     *ServerApi                     `inject:""`
}

func (h *Handler) Constructor() {
	h.server.Post(
		"/devices/:deviceId/:action",
		checkPath,
		convertCoreDeviceIdToRegistryMac,
		Action(h.service, h.repository),
	)

}

var cache sync.Map
var db *sqlx.DB

// TODO: Remove after demo
func convertCoreApiIdToRegistryMacId(deviceCoreId int) string {
	if item, hit := cache.Load(deviceCoreId); hit {
		return item.(string)
	}

	var macId []string

	if db == nil {
		var err error
		db, err = sqlx.Open("mysql", di.Application().GetParam("DATABASE_DSN"))
		if err != nil {
			panic(err)
		}
	}

	if err := db.Select(&macId, "SELECT mac_id FROM devices WHERE id = ?", deviceCoreId); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Println(err)
		}

		return ""
	}

	if len(macId) == 0 {
		return ""
	}

	return macId[0]
}
func convertCoreDeviceIdToRegistryMac(ctx *fiber.Ctx) error {
	var deviceId string
	coreDeviceId, err := strconv.Atoi(ctx.Params("deviceId", ""))

	if err == nil {
		deviceId = convertCoreApiIdToRegistryMacId(coreDeviceId)
	} else {
		deviceId = ctx.Params("deviceId", "")
	}

	// Hack to replace integer device id by macId, known by Registry API
	values := (*[30]string)(unsafe.Pointer(reflect.ValueNoEscapeOf(ctx).Elem().FieldByName("values").UnsafeAddr()))

	for index, name := range ctx.Route().Params {
		if name == "deviceId" {
			values[index] = deviceId
			break
		}
	}

	return ctx.Next()
}
func checkPath(ctx *fiber.Ctx) error {
	switch ctx.Params("action") {
	case "open":
	case "close":
	case "auto":
	default:
		return fiber.ErrNotFound
	}
	return ctx.Next()
}
