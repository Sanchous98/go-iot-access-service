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

	brokerRepository Repository[*models.Broker]            `inject:""`
	cache            cache.CacheInterface[*models.Gateway] `inject:""`
}

func (r *GatewayRepository) Find(id uuid.UUID) *models.Gateway {
	if item, err := r.cache.Get(context.Background(), id.String()); err == nil && item != nil {
		log.Printf("Gateway %s hitted cache\n", id.String())
		item.BrokerResolver = func() *models.Broker { return r.brokerRepository.Find(item.BrokerId) }
		return item
	}

	gateway := r.RegistryRepository.find(id)

	if gateway == nil {
		return nil
	}

	r.pushCache(gateway)

	gateway.BrokerResolver = func() *models.Broker { return r.brokerRepository.Find(gateway.BrokerId) }
	return gateway
}

func (r *GatewayRepository) FindByMacId(gatewayIeee string) (item *models.Gateway) {
	var err error

	if item, err = r.cache.Get(context.Background(), gatewayIeee); errors.Is(err, new(store.NotFound)) {
		gateways := r.RegistryRepository.findAll(map[string]any{"gatewayIeee": gatewayIeee})

		if len(gateways) == 0 {
			return nil
		}

		item = gateways[0]
		r.pushCache(item)
	} else if err != nil {
		log.Println(err)
		return nil
	}

	item.BrokerResolver = func() *models.Broker {
		return r.brokerRepository.Find(item.BrokerId)
	}

	return
}

func (r *GatewayRepository) FindAll() []*models.Gateway {
	gateways := r.RegistryRepository.findAll(map[string]any{"enabled": 1, "claimed": 1})

	if len(gateways) == 0 {
		return gateways
	}

	for _, gateway := range gateways {
		r.pushCache(gateway)

		gateway.BrokerResolver = func() *models.Broker { return r.brokerRepository.Find(gateway.BrokerId) }
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
}
