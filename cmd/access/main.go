package main

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/api"
	"bitbucket.org/4suites/iot-service-golang/pkg/cache"
	"bitbucket.org/4suites/iot-service-golang/pkg/handlers"
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"context"
	"github.com/Sanchous98/go-di"
	"github.com/allegro/bigcache/v3"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	gocache "github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/metrics"
	"github.com/eko/gocache/v3/store"
	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
	"time"
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
	app.AddEntryPoint(bootstrap)

	app.Set(new(api.ServerApi))
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

	app.Set(func(container di.Container) gocache.CacheInterface[*models.Broker] {
		return gocache.NewMetric[*models.Broker](promMetrics, cache.New[models.Broker](container.Get(new(store.StoreInterface)).(store.StoreInterface)))
	})
	app.Set(func(container di.Container) gocache.CacheInterface[*models.Device] {
		return gocache.NewMetric[*models.Device](promMetrics, cache.New[models.Device](container.Get(new(store.StoreInterface)).(store.StoreInterface)))
	})
	app.Set(func(container di.Container) gocache.CacheInterface[*models.Gateway] {
		return gocache.NewMetric[*models.Gateway](promMetrics, cache.New[models.Gateway](container.Get(new(store.StoreInterface)).(store.StoreInterface)))
	})

	app.Set(func(container di.Container) repositories.Repository[*models.Gateway] {
		return container.Build(new(repositories.GatewayRepository)).(*repositories.GatewayRepository)
	})

	app.Set(func(container di.Container) repositories.Repository[*models.Broker] {
		return container.Build(new(repositories.BrokerRepository)).(*repositories.BrokerRepository)
	})

	app.Set(new(services.HandlerAggregator[*models.Broker]))
	app.Set(new(services.HandlerAggregator[*models.Gateway]))

	app.Set(new(handlers.VerifyOnlineHandler), "mqtt.message_handler")
	app.Set(new(handlers.LocalStorageQueue), "mqtt.message_handler")

	app.Run(app.LoadEnv)
}

func bootstrap(container di.GlobalState) {
	if container.GetParam("APP_ENV") != "prod" && container.GetParam("APP_ENV") != "production" {
		mqtt.ERROR = log.New(os.Stdout, "[mqtt:ERROR]::", log.LUTC)
		mqtt.WARN = log.New(os.Stdout, "[mqtt:WARN]::", log.LUTC)
	}

	mqtt.CRITICAL = log.New(os.Stdout, "[mqtt:CRITICAL]::", log.LUTC)

	//if container.GetParam("APP_ENV") == "dev" || container.GetParam("APP_ENV") == "development" {
	//	mqtt.DEBUG = log.New(os.Stdout, "[mqtt:DEBUG]::", log.LUTC)
	//}

	profiler := fiber.New(fiber.Config{
		JSONDecoder: func(data []byte, v any) error { return json.UnmarshalNoEscape(data, v) },
		JSONEncoder: func(v any) ([]byte, error) { return json.MarshalNoEscape(v) },
	})
	profiler.Use(pprof.New())
	log.Println(profiler.Listen(":6060"))
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

		return nil
	}
}
