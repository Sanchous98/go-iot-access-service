package middleware

import (
	"github.com/goccy/go-reflect"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"log"
	"strconv"
	"sync"
	"unsafe"
)

var cache sync.Map

type device struct {
	Id    int    `gorm:"id"`
	MacId string `gorm:"mac_id"`
}

func (device) TableName() string {
	return "devices"
}

// TODO: Remove after demo
func convertCoreApiIdToRegistryMacId(db *gorm.DB, deviceCoreId int) string {
	if item, hit := cache.Load(deviceCoreId); hit {
		return item.(string)
	}

	d := device{Id: deviceCoreId}
	if err := db.First(&d, d).Error; err != nil {
		log.Println(err)
		return ""
	}

	return d.MacId
}
func ConvertCoreDeviceIdToRegistryMac(db *gorm.DB) fiber.Handler {
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
