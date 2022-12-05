package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"context"
	"errors"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/google/uuid"
	"log"
)

type DeviceRepository struct {
	RegistryRepository[*entities.Device] `inject:""`

	gatewayRepository application.Repository[*entities.Gateway] `inject:""`
	cache             cache.CacheInterface[*entities.Device]    `inject:""`
	brokerCache       cache.CacheInterface[*entities.Broker]    `inject:""`
	gatewayCache      cache.CacheInterface[*entities.Gateway]   `inject:""`
}

func (r *DeviceRepository) Find(id uuid.UUID) *entities.Device {
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

func (r *DeviceRepository) FindByMacId(macId string) (item *entities.Device) {
	var err error

	if item, err = r.cache.Get(context.Background(), macId); errors.Is(err, new(store.NotFound)) {
		item = r.FindOneBy(map[string]any{"macId": macId, "include": "gateway.broker"})
	} else if err != nil {
		log.Println(err)
		return nil
	}

	return
}

func (r *DeviceRepository) FindByMacIdAndGatewayIeee(deviceMacId string, gatewayIeee string) (item *entities.Device) {
	var err error

	if item, err = r.cache.Get(context.Background(), deviceMacId+gatewayIeee); errors.Is(err, new(store.NotFound)) {
		item = r.FindOneBy(map[string]any{"macId": deviceMacId, "include": "gateway.broker", "gateway.gatewayIeee": gatewayIeee})
	} else if err != nil {
		log.Println(err)
		return nil
	}

	return
}

func (r *DeviceRepository) FindAll() []*entities.Device {
	return r.FindBy(map[string]any{"enabled": 1, "claimed": 1, "include": "gateway.broker"})
}

func (r *DeviceRepository) FindOneBy(params map[string]any) *entities.Device {
	results := r.FindBy(params)

	if results == nil || len(results) == 0 {
		return nil
	}

	return results[0]
}

func (r *DeviceRepository) FindBy(params map[string]any) []*entities.Device {
	devices := r.RegistryRepository.findAll(params)

	if len(devices) == 0 {
		return devices
	}

	for _, device := range devices {
		r.pushCache(device)
	}

	return devices
}

func (r *DeviceRepository) pushCache(device *entities.Device) {
	if err := r.cache.Set(context.Background(), device.Id.String(), device); err != nil {
		log.Println(err)
	}

	if err := r.cache.Set(context.Background(), device.MacId, device); err != nil {
		log.Println(err)
	}

	if err := r.cache.Set(context.Background(), device.MacId+device.Gateway.GatewayIeee, device); err != nil {
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
