package models

import (
	"bitbucket.org/4suites/iot-service-golang/utils"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Device struct {
	Id                 utils.UUID `json:"id"`
	UserId             utils.UUID `json:"userId"`
	GatewayId          utils.UUID `json:"gatewayId"`
	TypeId             string     `json:"typeId"`
	FirmwareId         string     `json:"firmwareId"`
	Version            string     `json:"version"`
	MacId              string     `json:"macId"`
	TotalChannelsCount int        `json:"totalChannelsCount"`
	ClaimedChannels    []int      `json:"claimedChannels"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`

	GatewayResolver func() *Gateway
}

func (d *Device) GetEventsTopic() string {
	return fmt.Sprintf("$foursuites/gw/%s/dev/%s/events", d.GetGateway().GatewayIeee, d.MacId)
}

func (d *Device) GetCommandsTopic() string {
	return fmt.Sprintf("$foursuites/gw/%s/dev/%s/actions", d.GetGateway().GatewayIeee, d.MacId)
}
func (d *Device) GetOptions() *mqtt.ClientOptions { return d.GetGateway().GetOptions() }
func (d *Device) GetGateway() *Gateway            { return d.GatewayResolver() }
func (d *Device) GetEndpoint() string             { return "/locks" }
