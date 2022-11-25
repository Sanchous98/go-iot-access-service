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

type BrokerRepository struct {
	*RegistryRepository[*models.Broker] `inject:""`

	cache cache.SetterCacheInterface[*models.Broker] `inject:""`
}

func (r *BrokerRepository) Find(id uuid.UUID) *models.Broker {
	// TODO: Try to refactor
	if item, err := r.cache.GetCodec().Get(context.Background(), id.String()); err == nil && item != nil {
		log.Printf("Broker %s hitted cache\n", id.String())
		var broker models.Broker
		_ = json.UnmarshalNoEscape(utils.StrToBytes(item.(string)), &broker)
		return &broker
	}

	item := r.RegistryRepository.find(id)

	if item == nil {
		return nil
	}

	if err := r.cache.Set(context.Background(), id.String(), item, store.WithExpiration(1*time.Hour)); err != nil {
		log.Println(err)
	}

	return item
}

func (r *BrokerRepository) FindAll() []*models.Broker {
	brokers := r.RegistryRepository.findAll(map[string]any{"enabled": 1, "claimed": 1})

	for _, broker := range brokers {
		if err := r.cache.Set(context.Background(), broker.Id, broker, store.WithExpiration(1*time.Hour)); err != nil {
			log.Println(err)
		}
	}

	return brokers
}
