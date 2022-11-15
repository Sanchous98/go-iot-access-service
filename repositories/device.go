package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/models"
)

type DeviceRepository struct {
	*RegistryRepository[*models.Device] `inject:""`
	GatewayRepository                   Repository[*models.Gateway] `inject:""`
}

func (r *DeviceRepository) Find(id string) *models.Device {
	device := r.RegistryRepository.Find(id)

	if device == nil {
		return nil
	}

	device.GatewayResolver = func() *models.Gateway {
		return r.GatewayRepository.Find(device.GatewayId.String())
	}
	return device
}

func (r *DeviceRepository) FindByMacId(macId string) *models.Device {
	devices := r.RegistryRepository.findAll(map[string]any{"macId": macId})

	if len(devices) == 0 {
		return nil
	}

	item := devices[0]
	item.GatewayResolver = func() *models.Gateway {
		return r.GatewayRepository.Find(item.GatewayId.String())
	}

	return item
}

func (r *DeviceRepository) FindAll() []*models.Device {
	devices := r.RegistryRepository.FindAll()

	for _, device := range devices {
		device.GatewayResolver = func() *models.Gateway {
			return r.GatewayRepository.Find(device.GatewayId.String())
		}
	}

	return devices
}
