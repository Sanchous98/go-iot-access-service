package models

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type Gateway struct {
	Id uuid.UUID `json:"id"`
	//PublicId                   string         `json:"publicId"`
	//Version                    string         `json:"version"`
	//Channel                    int            `json:"channel"`
	//NetworkKey                 string         `json:"networkKey"`
	//UserId                     uuid.UUID     `json:"userId"`
	BrokerId uuid.UUID `json:"brokerId"`
	//SerialNumber               string         `json:"serialNumber"`
	//PanId                      string         `json:"panId"`
	//Hostname                   string         `json:"hostname"`
	//EthernetIeee               string         `json:"ethernetIeee"`
	GatewayIeee string `json:"gatewayIeee"`
	//VpnIp                      string         `json:"vpnIp"`
	//LocalIp                    string         `json:"localIp"`
	//HeartbeatTimeout           int            `json:"heartbeatTimeout"`
	//DoorDeviceHeartbeatTimeout int            `json:"doorDeviceHeartbeatTimeout"`
	//Token                      string         `json:"token"`
	//Metadata                   map[string]any `json:"metadata"`
	//Claimed                    bool           `json:"claimed"`
	//Synced                     bool           `json:"synced"`
	//CreatedAt                  time.Time      `json:"createdAt"`
	//UpdatedAt                  time.Time      `json:"updatedAt"`

	BrokerResolver func() *Broker `json:"-"`
}

func (g *Gateway) GetId() uuid.UUID { return g.Id }
func (g *Gateway) GetTopics() map[string]byte {
	return map[string]byte{
		fmt.Sprintf("$foursuites/gw/%s/info", g.GatewayIeee):          0,
		fmt.Sprintf("$foursuites/gw/%s/dev/+/events", g.GatewayIeee):  0,
		fmt.Sprintf("$foursuites/gw/%s/dev/+/actions", g.GatewayIeee): 0,
	}
}
func (g *Gateway) GetOptions() *mqtt.ClientOptions { return g.GetBroker().GetOptions() }
func (g *Gateway) GetBroker() *Broker              { return g.BrokerResolver() }
func (*Gateway) GetResource() string               { return "gateways" }

func (g *Gateway) UnmarshalBinary(data []byte) error       { return json.UnmarshalNoEscape(data, g) }
func (g *Gateway) MarshalBinary() (data []byte, err error) { return json.MarshalNoEscape(g) }
