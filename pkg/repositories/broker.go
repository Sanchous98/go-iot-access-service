package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/cache"
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"github.com/google/uuid"
)

type BrokerRepository struct {
	*RegistryRepository[*models.Broker] `inject:""`

	cache cache.Cache[*models.Broker]
}

func (r *BrokerRepository) Find(id uuid.UUID) *models.Broker {
	if item, hit := r.cache.Get(func(b *models.Broker) bool { return b.Id == id }); hit {
		return item
	}

	item := r.RegistryRepository.Find(id)

	if item == nil {
		return nil
	}

	r.cache.Put(item)
	return item
}

func (r *BrokerRepository) FindAll() []*models.Broker {
	brokers := r.RegistryRepository.FindAll()
	r.cache.Put(brokers...)

	return brokers
}
