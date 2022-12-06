package cache

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/unsafe"
	"context"
	"github.com/Sanchous98/go-di"
	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/lib/v4/store"
	bigcacheStore "github.com/eko/gocache/store/bigcache/v4"
	redisStore "github.com/eko/gocache/store/redis/v4"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

func Factory(ctx context.Context) func(environment di.GlobalState) store.StoreInterface {
	return func(environment di.GlobalState) store.StoreInterface {
		switch environment.GetParam("CACHE_STORAGE") {
		case "null", "":
			return new(NullStore)
		case "memory":
			return bigcacheStore.NewBigcache(unsafe.Must(bigcache.New(ctx, bigcache.DefaultConfig(5*time.Minute))))
		case "redis":
			return redisStore.NewRedis(redis.NewClient(&redis.Options{
				Addr: environment.GetParam("REDIS_HOST"),
				DB:   unsafe.Must(strconv.Atoi(environment.GetParam("REDIS_DB"))),
			}), store.WithExpiration(1*time.Hour))
		}

		panic("unknown cache storage")
	}
}
