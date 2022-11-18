package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/cache"
	"bitbucket.org/4suites/iot-service-golang/models"
	"bitbucket.org/4suites/iot-service-golang/utils"
)

type DeviceRepository struct {
	*RegistryRepository[*models.Device] `inject:""`
	GatewayRepository                   Repository[*models.Gateway] `inject:""`
	cache                               cache.Cache[*models.Device]
}

func (r *DeviceRepository) Find(id utils.UUID) *models.Device {
	if item, hit := r.cache.Get(func(d *models.Device) bool { return d.Id == id }); hit {
		return item
	}

	device := r.RegistryRepository.Find(id)

	if device == nil {
		return nil
	}

	r.cache.Put(device)

	device.GatewayResolver = func() *models.Gateway {
		return r.GatewayRepository.Find(device.GatewayId.String())
	}
	return device
}

func (r *DeviceRepository) FindByMacId(macId string) *models.Device {
	if item, hit := r.cache.Get(func(d *models.Device) bool { return d.MacId == macId }); hit {
		return item
	}

	devices := r.RegistryRepository.findAll(map[string]any{"macId": macId})

	if len(devices) == 0 {
		return nil
	}

	item := devices[0]
	r.cache.Put(item)
	item.GatewayResolver = func() *models.Gateway {
		return r.GatewayRepository.Find(item.GatewayId.String())
	}

	return item
}

func (r *DeviceRepository) FindAll() []*models.Device {
	devices := r.RegistryRepository.FindAll()

	if len(devices) == 0 {
		return devices
	}

	r.cache.Put(devices...)

	for _, device := range devices {
		device.GatewayResolver = func() *models.Gateway {
			return r.GatewayRepository.Find(device.GatewayId.String())
		}
	}

	return devices
}
