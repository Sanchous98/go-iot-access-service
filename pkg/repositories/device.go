package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/models"
	"context"
	"github.com/eko/gocache/v3/cache"
	"github.com/google/uuid"
	"log"
)

type DeviceRepository struct {
	RegistryRepository[*models.Device] `inject:""`

	gatewayRepository Repository[*models.Gateway]           `inject:""`
	cache             cache.CacheInterface[*models.Device]  `inject:""`
	brokerCache       cache.CacheInterface[*models.Broker]  `inject:""`
	gatewayCache      cache.CacheInterface[*models.Gateway] `inject:""`
}

func (r *DeviceRepository) Find(id uuid.UUID) *models.Device {
	if item, err := r.cache.Get(context.Background(), id.String()); err == nil && item != nil {
		log.Printf("Device %s hitted cache\n", id.String())
		return item
	}

	device := r.RegistryRepository.find(id, map[string]any{"include": "gateway.broker"})

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

func (r *DeviceRepository) FindByMacIdAndGatewayIeee(deviceMacId string, gatewayIeee string) *models.Device {
	if item, err := r.cache.Get(context.Background(), deviceMacId); err == nil && item != nil {
		log.Printf("Device %s hitted cache\n", deviceMacId)
		return item
	}

	devices := r.RegistryRepository.findAll(map[string]any{"macId": deviceMacId, "include": "gateway.broker", "gateway.gatewayIeee": gatewayIeee})

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

	if err := r.gatewayCache.Set(context.Background(), device.GatewayId.String(), device.Gateway); err != nil {
		log.Println(err)
	}

	if err := r.gatewayCache.Set(context.Background(), device.Gateway.GatewayIeee, device.Gateway); err != nil {
		log.Println(err)
	}

	if err := r.brokerCache.Set(context.Background(), device.Gateway.BrokerId.String(), device.Gateway.Broker); err != nil {
		log.Println(err)
	}
}
