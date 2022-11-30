package main

import (
	"bitbucket.org/4suites/iot-service-golang/access/api"
	"bitbucket.org/4suites/iot-service-golang/cmd/internal"
	"bitbucket.org/4suites/iot-service-golang/pkg/cache"
	"bitbucket.org/4suites/iot-service-golang/pkg/http"
	"bitbucket.org/4suites/iot-service-golang/pkg/listeners"
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"bitbucket.org/4suites/iot-service-golang/pkg/repositories"
	"bitbucket.org/4suites/iot-service-golang/pkg/services"
	"context"
	"github.com/Sanchous98/go-di"
	"github.com/allegro/bigcache/v3"
	gocache "github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/metrics"
	"github.com/eko/gocache/v3/store"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
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

	app.Set(new(services.HandlerAggregator))
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
