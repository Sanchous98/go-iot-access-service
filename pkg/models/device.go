package models

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type Device struct {
	Id uuid.UUID `json:"id"`
	//UserId             uuid.UUID `json:"userId"`
	GatewayId uuid.UUID `json:"gatewayId"`
	//TypeId             string     `json:"typeId"`
	//FirmwareId         string     `json:"firmwareId"`
	//Version            string     `json:"version"`
	MacId string `json:"macId"`
	//TotalChannelsCount int        `json:"totalChannelsCount"`
	//ClaimedChannels    []int      `json:"claimedChannels"`
	//CreatedAt          time.Time  `json:"createdAt"`
	//UpdatedAt          time.Time  `json:"updatedAt"`

	GatewayResolver func() *Gateway `json:"-"`
}

func (d *Device) GetEventsTopic() string {
	return fmt.Sprintf("$foursuites/gw/%s/dev/%s/events", d.GetGateway().GatewayIeee, d.MacId)
}

func (d *Device) GetCommandsTopic() string {
	return fmt.Sprintf("$foursuites/gw/%s/dev/%s/actions", d.GetGateway().GatewayIeee, d.MacId)
}
func (d *Device) GetOptions() *mqtt.ClientOptions { return d.GetGateway().GetOptions() }
func (d *Device) GetGateway() *Gateway            { return d.GatewayResolver() }
func (*Device) GetResource() string               { return "locks" }

func (d *Device) UnmarshalBinary(data []byte) error       { return json.UnmarshalNoEscape(data, d) }
func (d *Device) MarshalBinary() (data []byte, err error) { return json.MarshalNoEscape(d) }
