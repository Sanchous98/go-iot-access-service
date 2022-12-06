package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"context"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/google/uuid"
)

type BrokerRepository struct {
	RegistryRepository[*entities.Broker] `inject:""`

	cache cache.CacheInterface[*entities.Broker] `inject:""`

	log logger.Logger `inject:""`
}

func (r *BrokerRepository) Find(id uuid.UUID) (item *entities.Broker) {
	var err error

	if item, err = r.cache.Get(context.Background(), id.String()); err != nil {
		item = r.RegistryRepository.find(id, map[string]any{"include": "gateway.broker"})

		if item == nil {
			r.log.Debugf("Device %s not found\n", id.String())
			return
		}

		r.pushCache(item)
	} else {
		r.log.Infof("Device %s hit cache\n", id.String())
	}

	return item
}

func (r *BrokerRepository) FindAll() []*entities.Broker {
	return r.FindBy(map[string]any{"enabled": 1, "claimed": 1})
}

func (r *BrokerRepository) FindOneBy(params map[string]any) *entities.Broker {
	results := r.FindBy(params)

	if results == nil || len(results) == 0 {
		return nil
	}

	return results[0]
}

func (r *BrokerRepository) FindBy(params map[string]any) []*entities.Broker {
	brokers := r.RegistryRepository.findAll(params)

	for _, broker := range brokers {
		r.pushCache(broker)
	}

	return brokers
}

func (r *BrokerRepository) pushCache(broker *entities.Broker) {
	if err := r.cache.Set(context.Background(), broker.Id, broker); err != nil {
		r.log.Errorln(err)
	}
}
