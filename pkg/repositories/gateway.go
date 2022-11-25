package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"bitbucket.org/4suites/iot-service-golang/pkg/utils"
	"context"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"log"
	"time"
)

type GatewayRepository struct {
	*RegistryRepository[*models.Gateway] `inject:""`

	brokerRepository Repository[*models.Broker]                  `inject:""`
	cache            cache.SetterCacheInterface[*models.Gateway] `inject:""`
}

func (r *GatewayRepository) Find(id uuid.UUID) *models.Gateway {
	if item, err := r.cache.GetCodec().Get(context.Background(), id.String()); err == nil && item != nil {
		log.Printf("Gateway %s hitted cache\n", id.String())
		var gateway models.Gateway
		_ = json.UnmarshalNoEscape(utils.StrToBytes(item.(string)), &gateway)
		gateway.BrokerResolver = func() *models.Broker { return r.brokerRepository.Find(gateway.BrokerId) }
		return &gateway
	}

	gateway := r.RegistryRepository.find(id)

	if gateway == nil {
		return nil
	}

	r.pushCache(gateway)

	gateway.BrokerResolver = func() *models.Broker { return r.brokerRepository.Find(gateway.BrokerId) }
	return gateway
}

func (r *GatewayRepository) FindByMacId(macId string) *models.Gateway {
	// TODO: Try to refactor
	if item, err := r.cache.GetCodec().Get(context.Background(), macId); err == nil && item != nil {
		log.Printf("Device %s hitted cache\n", macId)
		var gateway models.Gateway
		_ = json.UnmarshalNoEscape(utils.StrToBytes(item.(string)), &gateway)
		gateway.BrokerResolver = func() *models.Broker { return r.brokerRepository.Find(gateway.BrokerId) }
		return &gateway
	}

	gateways := r.RegistryRepository.findAll(map[string]any{"gatewayIeee": macId})

	if len(gateways) == 0 {
		return nil
	}

	item := gateways[0]
	r.pushCache(item)

	item.BrokerResolver = func() *models.Broker {
		return r.brokerRepository.Find(item.BrokerId)
	}

	return item
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
	if err := r.cache.Set(context.Background(), gateway.Id.String(), gateway, store.WithExpiration(1*time.Hour)); err != nil {
		log.Println(err)
	}

	if err := r.cache.Set(context.Background(), gateway.GatewayIeee, gateway, store.WithExpiration(1*time.Hour)); err != nil {
		log.Println(err)
	}
}
