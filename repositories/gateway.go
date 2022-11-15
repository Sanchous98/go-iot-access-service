package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/models"
)

type GatewayRepository struct {
	*RegistryRepository[*models.Gateway] `inject:""`
	BrokerRepository                     Repository[*models.Broker] `inject:""`
}

func (r *GatewayRepository) Find(id string) *models.Gateway {
	gateway := r.RegistryRepository.Find(id)

	if gateway == nil {
		return nil
	}

	gateway.BrokerResolver = func() *models.Broker {
		return r.BrokerRepository.Find(gateway.BrokerId.String())
	}
	return gateway
}

func (r *GatewayRepository) FindByMacId(macId string) *models.Gateway {
	gateways := r.RegistryRepository.findAll(map[string]any{"gatewayIeee": macId})

	if len(gateways) == 0 {
		return nil
	}

	item := gateways[0]
	item.BrokerResolver = func() *models.Broker {
		return r.BrokerRepository.Find(item.BrokerId.String())
	}

	return item
}

func (r *GatewayRepository) FindAll() []*models.Gateway {
	gateways := r.RegistryRepository.FindAll()

	for _, gateway := range gateways {
		gateway.BrokerResolver = func() *models.Broker {
			return r.BrokerRepository.Find(gateway.BrokerId.String())
		}
	}

	return gateways
}
