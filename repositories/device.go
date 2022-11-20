package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/cache"
	"bitbucket.org/4suites/iot-service-golang/models"
	"github.com/google/uuid"
)

type DeviceRepository struct {
	*RegistryRepository[*models.Device] `inject:""`
	GatewayRepository                   Repository[*models.Gateway] `inject:""`
	BrokerRepository                    Repository[*models.Broker]  `inject:""`
	cache                               cache.Cache[*models.Device]
}

func (r *DeviceRepository) Find(id uuid.UUID) *models.Device {
	if item, hit := r.cache.Get(func(d *models.Device) bool { return d.Id == id }); hit {
		return item
	}

	device := r.RegistryRepository.Find(id)

	if device == nil {
		return nil
	}

	r.cache.Put(device)

	device.GatewayResolver = func() *models.Gateway {
		return r.GatewayRepository.Find(device.GatewayId)
	}
	return device
}

func (r *DeviceRepository) FindByMacId(macId string) *models.Device {
	if macId == "0x0000000001" {
		broker := r.BrokerRepository.Find(uuid.MustParse("07abfaac-321d-43f7-bf00-8f27f44199f1"))

		gateway := &models.Gateway{
			Id:          uuid.New(),
			GatewayIeee: macId,
			BrokerId:    broker.Id,
			BrokerResolver: func() *models.Broker {
				return broker
			},
		}

		return &models.Device{
			Id:        uuid.New(),
			GatewayId: gateway.Id,
			MacId:     macId,
			GatewayResolver: func() *models.Gateway {
				return gateway
			},
		}
	}

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
		return r.GatewayRepository.Find(item.GatewayId)
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
			return r.GatewayRepository.Find(device.GatewayId)
		}
	}

	return devices
}
