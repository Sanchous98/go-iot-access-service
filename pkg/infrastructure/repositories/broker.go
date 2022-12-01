package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"context"
	"github.com/eko/gocache/v3/cache"
	"github.com/google/uuid"
	"log"
)

type BrokerRepository struct {
	RegistryRepository[*entities.Broker] `inject:""`

	cache cache.CacheInterface[*entities.Broker] `inject:""`
}

func (r *BrokerRepository) Find(id uuid.UUID) *entities.Broker {
	if item, err := r.cache.Get(context.Background(), id.String()); err == nil && item != nil {
		log.Printf("Broker %s hitted cache\n", id.String())
		return item
	}

	item := r.RegistryRepository.find(id, nil)

	if item == nil {
		return nil
	}

	if err := r.cache.Set(context.Background(), id.String(), item); err != nil {
		log.Println(err)
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
		if err := r.cache.Set(context.Background(), broker.Id, broker); err != nil {
			log.Println(err)
		}
	}

	return brokers
}
