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

	return device
}

func (r *DeviceRepository) FindByMacId(macId string) *models.Device {
	if item, err := r.cache.Get(context.Background(), macId); err == nil && item != nil {
		log.Printf("Device %s hitted cache\n", macId)
		return item
	}

	devices := r.RegistryRepository.findAll(map[string]any{"macId": macId, "include": "gateway.broker"})

	if len(devices) == 0 {
		return nil
	}

	item := devices[0]
	r.pushCache(item)

	return item
}

func (r *DeviceRepository) FindAll() []*models.Device {
	devices := r.RegistryRepository.findAll(map[string]any{"enabled": 1, "claimed": 1, "include": "gateway.broker"})

	if len(devices) == 0 {
		return devices
	}

	for _, device := range devices {
		r.pushCache(device)
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
