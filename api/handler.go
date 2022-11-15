package api

import (
	"bitbucket.org/4suites/iot-service-golang/repositories"
	"bitbucket.org/4suites/iot-service-golang/services"
)

type Handler struct {
	service    services.DeviceService         `inject:""`
	repository *repositories.DeviceRepository `inject:""`
	server     *ServerApi                     `inject:""`
}

func (h *Handler) Constructor() {
	h.server.Post("/devices/:deviceId/open", Open(h.service, h.repository))
	h.server.Post("/devices/:deviceId/close", Close(h.service, h.repository))
	h.server.Post("/devices/:deviceId/auto", Auto(h.service, h.repository))
}
