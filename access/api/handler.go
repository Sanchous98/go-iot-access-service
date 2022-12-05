package api

import (
	"bitbucket.org/4suites/iot-service-golang/access/api/controllers"
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/application/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/http/middleware"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/repositories"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AccessApiHandler struct {
	service    services.DeviceService        `inject:""`
	repository repositories.DeviceRepository `inject:""`
	db         *gorm.DB                      `inject:""`

	brokerRepository repositories.BrokerRepository          `inject:""`
	brokerCache      cache.CacheInterface[*entities.Broker] `inject:""`
	pool             application.HandlerPool                `inject:""`
}

func (h *AccessApiHandler) RegisterRoutes(app *fiber.App) {
	app.Post(
		"/devices/:deviceId/:action",
		checkPath,
		middleware.ConvertCoreDeviceIdToRegistryMac(h.db),
		controllers.Action(h.service, h.repository),
	)

	app.Post("/broker/:brokerId/disconnect", controllers.DisconnectBroker(h.brokerRepository, h.brokerCache, h.pool))
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
