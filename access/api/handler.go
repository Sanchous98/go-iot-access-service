package api

import (
	"bitbucket.org/4suites/iot-service-golang/access/api/controllers"
	"bitbucket.org/4suites/iot-service-golang/pkg/http/middleware"
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AccessApiHandler struct {
	service    services.DeviceService        `inject:""`
	repository repositories.DeviceRepository `inject:""`
	db         *gorm.DB                      `inject:""`
}

func (h *AccessApiHandler) RegisterRoutes(app *fiber.App) {
	app.Post(
		"/devices/:deviceId/:action",
		checkPath,
		middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
		controllers.Action(h.service, h.repository),
	)
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
