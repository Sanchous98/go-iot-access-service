package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/models"
)

type BrokerRepository struct {
	*RegistryRepository[*models.Broker] `inject:""`
}
