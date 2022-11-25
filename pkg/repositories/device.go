package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"context"
	"github.com/eko/gocache/v3/cache"
	"github.com/google/uuid"
	"log"
)

type DeviceRepository struct {
	*RegistryRepository[*models.Device] `inject:""`

	gatewayRepository Repository[*models.Gateway]          `inject:""`
	cache             cache.CacheInterface[*models.Device] `inject:""`
}

func (r *DeviceRepository) Find(id uuid.UUID) *models.Device {
	if item, err := r.cache.Get(context.Background(), id.String()); err == nil && item != nil {
		log.Printf("Device %s hitted cache\n", id.String())
		return item
	}

	device := r.RegistryRepository.find(id)

	if device == nil {
		return nil
	}

	r.pushCache(device)
	device.GatewayResolver = func() *models.Gateway { return r.gatewayRepository.Find(device.GatewayId) }

	return device
}

func (r *DeviceRepository) FindByMacId(macId string) *models.Device {
	// TODO: Try to refactor
	if item, err := r.cache.Get(context.Background(), macId); err == nil && item != nil {
		log.Printf("Device %s hitted cache\n", macId)
		item.GatewayResolver = func() *models.Gateway {
			return r.gatewayRepository.Find(item.GatewayId)
		}
		return item
	}

	devices := r.RegistryRepository.findAll(map[string]any{"macId": macId})

	if len(devices) == 0 {
		return nil
	}

	item := devices[0]
	r.pushCache(item)

	item.GatewayResolver = func() *models.Gateway {
		return r.gatewayRepository.Find(item.GatewayId)
	}

	return item
}

func (r *DeviceRepository) FindAll() []*models.Device {
	devices := r.RegistryRepository.findAll(map[string]any{"enabled": 1, "claimed": 1})

	if len(devices) == 0 {
		return devices
	}

	for _, device := range devices {
		r.pushCache(device)

		device.GatewayResolver = func() *models.Gateway {
			return r.gatewayRepository.Find(device.GatewayId)
		}
	}

	return devices
}

func (r *DeviceRepository) pushCache(device *models.Device) {
	if err := r.cache.Set(context.Background(), device.Id.String(), device); err != nil {
		log.Println(err)
	}

	if err := r.cache.Set(context.Background(), device.MacId, device); err != nil {
		log.Println(err)
	}
}
