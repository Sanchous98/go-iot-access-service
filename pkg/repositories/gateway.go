package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/cache"
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"github.com/google/uuid"
)

type GatewayRepository struct {
	*RegistryRepository[*models.Gateway] `inject:""`

	brokerRepository Repository[*models.Broker] `inject:""`
	cache            cache.Cache[*models.Gateway]
}

func (r *GatewayRepository) Find(id uuid.UUID) *models.Gateway {
	if item, hit := r.cache.Get(func(g *models.Gateway) bool { return g.Id == id }); hit {
		return item
	}

	gateway := r.RegistryRepository.Find(id)

	if gateway == nil {
		return nil
	}

	gateway.BrokerResolver = func() *models.Broker {
		return r.brokerRepository.Find(gateway.BrokerId)
	}
	return gateway
}

func (r *GatewayRepository) FindByMacId(macId string) *models.Gateway {
	if item, hit := r.cache.Get(func(g *models.Gateway) bool { return g.GatewayIeee == macId }); hit {
		return item
	}

	gateways := r.RegistryRepository.findAll(map[string]any{"gatewayIeee": macId})

	if len(gateways) == 0 {
		return nil
	}

	item := gateways[0]
	item.BrokerResolver = func() *models.Broker {
		return r.brokerRepository.Find(item.BrokerId)
	}

	return item
}

func (r *GatewayRepository) FindAll() []*models.Gateway {
	gateways := r.RegistryRepository.FindAll()

	if len(gateways) == 0 {
		return gateways
	}

	r.cache.Put(gateways...)

	for _, gateway := range gateways {
		gateway.BrokerResolver = func() *models.Broker {
			return r.brokerRepository.Find(gateway.BrokerId)
		}
	}

	return gateways
}
