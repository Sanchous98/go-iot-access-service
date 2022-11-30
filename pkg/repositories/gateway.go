package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"context"
	"errors"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/google/uuid"
	"log"
)

type GatewayRepository struct {
	*RegistryRepository[*models.Gateway] `inject:""`

	cache       cache.CacheInterface[*models.Gateway] `inject:""`
	brokerCache cache.CacheInterface[*models.Broker]  `inject:""`
}

func (r *GatewayRepository) Find(id uuid.UUID) *models.Gateway {
	if item, err := r.cache.Get(context.Background(), id.String()); err == nil && item != nil {
		log.Printf("Gateway %s hitted cache\n", id.String())
		return item
	}

	gateway := r.RegistryRepository.find(id, map[string]any{"include": "broker"})

	if gateway == nil {
		return nil
	}

	r.pushCache(gateway)

	return gateway
}

func (r *GatewayRepository) FindByIeee(gatewayIeee string) (item *models.Gateway) {
	var err error

	if item, err = r.cache.Get(context.Background(), gatewayIeee); errors.Is(err, new(store.NotFound)) {
		gateways := r.RegistryRepository.findAll(map[string]any{"gatewayIeee": gatewayIeee, "include": "broker"})

		if len(gateways) == 0 {
			return nil
		}

		item = gateways[0]
		r.pushCache(item)
	} else if err != nil {
		log.Println(err)
		return nil
	}

	return
}

func (r *GatewayRepository) FindAll() []*models.Gateway {
	gateways := r.RegistryRepository.findAll(map[string]any{"enabled": 1, "claimed": 1, "include": "broker"})

	if len(gateways) == 0 {
		return gateways
	}

	for _, gateway := range gateways {
		r.pushCache(gateway)
	}

	return gateways
}

func (r *GatewayRepository) pushCache(gateway *models.Gateway) {
	if err := r.cache.Set(context.Background(), gateway.Id.String(), gateway); err != nil {
		log.Println(err)
	}

	if err := r.cache.Set(context.Background(), gateway.GatewayIeee, gateway); err != nil {
		log.Println(err)
	}

	if err := r.brokerCache.Set(context.Background(), gateway.BrokerId.String(), gateway.Broker); err != nil {
		log.Println(err)
	}
}
