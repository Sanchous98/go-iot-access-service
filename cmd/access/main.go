package main

import (
	"bitbucket.org/4suites/iot-service-golang/access/api"
	"bitbucket.org/4suites/iot-service-golang/cmd/internal"
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/application/listeners"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/cache"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/http"
	loggerWrapper "bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/logger"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/services"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/unsafe"
	"context"
	"github.com/Sanchous98/go-di"
	gocache "github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/metrics"
	"github.com/eko/gocache/lib/v4/store"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if err := recover(); err != nil {
			log.Printf("%#v", err)
			cancel()
		}
	}()

	app := di.Application(ctx)
	app.AddEntryPoint(internal.Bootstrap)

	app.Set(loggerWrapper.Factory)

	app.Set(new(http.ServerApi))
	app.Set(new(api.AccessApiHandler), "api.handler")
	app.Set(func(environment di.GlobalState) *gorm.DB {
		return unsafe.Must(gorm.Open(mysql.Open(environment.GetParam("DATABASE_DSN"))))
	})
	app.Set(cache.Factory(ctx))

	promMetrics := metrics.NewPrometheus("go-iot-access-service")

	app.Set(func(container di.Container) gocache.CacheInterface[*entities.Broker] {
		return gocache.NewMetric[*entities.Broker](promMetrics, cache.New[entities.Broker](container.Get(new(store.StoreInterface)).(store.StoreInterface)))
	})
	app.Set(func(container di.Container) gocache.CacheInterface[*entities.Device] {
		return gocache.NewMetric[*entities.Device](promMetrics, cache.New[entities.Device](container.Get(new(store.StoreInterface)).(store.StoreInterface)))
	})
	app.Set(func(container di.Container) gocache.CacheInterface[*entities.Gateway] {
		return gocache.NewMetric[*entities.Gateway](promMetrics, cache.New[entities.Gateway](container.Get(new(store.StoreInterface)).(store.StoreInterface)))
	})

	app.Set(new(repositories.GatewayRepository))
	app.Set(new(repositories.BrokerRepository))
	app.Set(new(repositories.DeviceRepository))

	app.Set(func(container di.Container) application.Repository[*entities.Device] {
		return container.Get((*repositories.DeviceRepository)(nil)).(*repositories.DeviceRepository)
	})
	app.Set(func(container di.Container) application.Repository[*entities.Broker] {
		return container.Get((*repositories.BrokerRepository)(nil)).(*repositories.BrokerRepository)
	})
	app.Set(func(container di.Container) application.Repository[*entities.Gateway] {
		return container.Get((*repositories.GatewayRepository)(nil)).(*repositories.GatewayRepository)
	})
	app.Set(func(container di.Container) application.DeviceRepository {
		return container.Get((*repositories.DeviceRepository)(nil)).(*repositories.DeviceRepository)
	})
	app.Set(func(container di.Container) application.GatewayRepository {
		return container.Get((*repositories.GatewayRepository)(nil)).(*repositories.GatewayRepository)
	})
	app.Set(new(services.MessageAggregator))
	app.Set(func(container di.Container) application.ClientPool {
		return container.Get((*services.MessageAggregator)(nil)).(application.ClientPool)
	})
	app.Set(new(listeners.VerifyOnlineHandler), "mqtt.message_handler")
	app.Run(app.LoadEnv)
}
