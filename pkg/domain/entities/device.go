package entities

import (
	"fmt"
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

	Gateway *Gateway `json:"gateway,ommitempty"`
}

func (d *Device) GetEventsTopic() string {
	return fmt.Sprintf("$foursuites/gw/%s/dev/%s/events", d.Gateway.GatewayIeee, d.MacId)
}

func (d *Device) GetCommandsTopic() string {
	return fmt.Sprintf("$foursuites/gw/%s/dev/%s/actions", d.Gateway.GatewayIeee, d.MacId)
}
func (*Device) GetResource() string { return "locks" }
