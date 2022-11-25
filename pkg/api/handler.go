package api

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"github.com/goccy/go-reflect"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"strconv"
	"sync"
	"unsafe"
)

var cache sync.Map

type Device struct {
	Id    int    `gorm:"id"`
	MacId string `gorm:"mac_id"`
}

func (Device) TableName() string {
	return "devices"
}

type Handler struct {
	service    services.DeviceService        `inject:""`
	repository repositories.DeviceRepository `inject:""`
	server     *ServerApi                    `inject:""`
	db         *gorm.DB                      `inject:""`
}

func (h *Handler) Constructor() {
	h.server.Post(
		"/devices/:deviceId/:action",
		checkPath,
		convertCoreDeviceIdToRegistryMac(h.db),
		Action(h.service, h.repository),
	)
}

// TODO: Remove after demo
func convertCoreApiIdToRegistryMacId(db *gorm.DB, deviceCoreId int) string {
	if item, hit := cache.Load(deviceCoreId); hit {
		return item.(string)
	}

	device := Device{Id: deviceCoreId}
	if err := db.First(&device, device).Error; err != nil {
		log.Println(err)
		return ""
	}

	return device.MacId
}
func convertCoreDeviceIdToRegistryMac(db *gorm.DB) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var deviceId string
		coreDeviceId, err := strconv.Atoi(ctx.Params("deviceId", ""))

		if err == nil {
			deviceId = convertCoreApiIdToRegistryMacId(db, coreDeviceId)
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
