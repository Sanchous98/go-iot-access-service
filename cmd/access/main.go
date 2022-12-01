package main

import (
	"bitbucket.org/4suites/iot-service-golang/access/api"
	"bitbucket.org/4suites/iot-service-golang/cmd/internal"
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/application/listeners"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/cache"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/http"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/services"
	"context"
	"log"
	"strconv"
	"time"

	"github.com/Sanchous98/go-di"
	"github.com/allegro/bigcache/v3"
	gocache "github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/metrics"
	"github.com/eko/gocache/v3/store"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if err := recover(); err != nil {
			log.Printf("%v", err)
			cancel()
		}
	}()

	app := di.Application(ctx)
	app.AddEntryPoint(internal.Bootstrap)

	app.Set(new(http.ServerApi))
	app.Set(new(api.AccessApiHandler), "api.handler")
	app.Set(func(environment di.GlobalState) *gorm.DB {
		db, err := gorm.Open(mysql.Open(environment.GetParam("DATABASE_DSN")))

		if err != nil {
			panic(err)
		}

		return db
	})
	app.Set(cacheFactory(ctx))

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
	app.Set(func(container di.Container) application.HandlerPool {
		return container.Build(new(services.HandlerAggregator)).(*services.HandlerAggregator)
	})
	app.Set(new(listeners.VerifyOnlineHandler), "mqtt.message_handler")

	app.Run(app.LoadEnv)
}

func cacheFactory(ctx context.Context) func(environment di.GlobalState) store.StoreInterface {
	return func(environment di.GlobalState) store.StoreInterface {
		switch environment.GetParam("CACHE_STORAGE") {
		case "null", "":
			return new(cache.NullStore)
		case "memory":
			db, err := bigcache.New(ctx, bigcache.DefaultConfig(5*time.Minute))

			if err != nil {
				panic(err)
			}

			return store.NewBigcache(db)
		case "redis":
			db, err := strconv.Atoi(environment.GetParam("REDIS_DB"))

			if err != nil {
				panic(err)
			}

			return store.NewRedis(redis.NewClient(&redis.Options{
				Addr: environment.GetParam("REDIS_HOST"),
				DB:   db,
			}), store.WithExpiration(1*time.Hour))
		}

		panic("unknown cache storage")
	}
}
