package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"context"
	"errors"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/google/uuid"
)

type GatewayRepository struct {
	RegistryRepository[*entities.Gateway] `inject:""`

	cache       cache.CacheInterface[*entities.Gateway] `inject:""`
	brokerCache cache.CacheInterface[*entities.Broker]  `inject:""`

	log logger.Logger `inject:""`
}

func (r *GatewayRepository) Find(id uuid.UUID) *entities.Gateway {
	if item, err := r.cache.Get(context.Background(), id.String()); err == nil && item != nil {
		r.log.Infof("Gateway %s hit cache\n", id.String())
		return item
	}

	gateway := r.RegistryRepository.find(id, map[string]any{"include": "broker"})

	if gateway == nil {
		return nil
	}

	r.pushCache(gateway)

	return gateway
}

func (r *GatewayRepository) FindByIeee(gatewayIeee string) (item *entities.Gateway) {
	var err error

	if item, err = r.cache.Get(context.Background(), gatewayIeee); !errors.Is(err, new(store.NotFound)) {
		item = r.FindOneBy(map[string]any{"gatewayIeee": gatewayIeee, "include": "broker"})
	} else if err != nil {
		r.log.Errorln(err)
		return nil
	}

	return
}

func (r *GatewayRepository) FindAll() []*entities.Gateway {
	return r.FindBy(map[string]any{"enabled": 1, "claimed": 1, "include": "broker"})
}

func (r *GatewayRepository) FindOneBy(params map[string]any) *entities.Gateway {
	results := r.FindBy(params)

	if results == nil || len(results) == 0 {
		return nil
	}

	return results[0]
}

func (r *GatewayRepository) FindBy(params map[string]any) []*entities.Gateway {
	gateways := r.RegistryRepository.findAll(params)

	for _, device := range gateways {
		r.pushCache(device)
	}

	return gateways
}

func (r *GatewayRepository) pushCache(gateway *entities.Gateway) {
	if err := r.cache.Set(context.Background(), gateway.Id.String(), gateway); err != nil {
		r.log.Errorln(err)
	}

	if err := r.cache.Set(context.Background(), gateway.GatewayIeee, gateway); err != nil {
		r.log.Errorln(err)
	}

	if err := r.brokerCache.Set(context.Background(), gateway.BrokerId.String(), gateway.Broker); err != nil {
		r.log.Errorln(err)
	}
}
