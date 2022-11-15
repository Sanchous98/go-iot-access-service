package models

import (
	"bitbucket.org/4suites/iot-service-golang/utils"
	"crypto/tls"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Broker struct {
	Id                utils.UUID     `json:"id"`
	UserId            utils.UUID     `json:"userId"`
	Name              string         `json:"name"`
	Host              string         `json:"host"`
	Port              int            `json:"port"`
	Claimed           bool           `json:"claimed"`
	Enabled           bool           `json:"enabled"`
	Metadata          map[string]any `json:"metadata"`
	CaCertificate     string         `json:"caCertificate"`
	ClientCertificate string         `json:"clientCertificate"`
	ClientKey         string         `json:"clientKey"`
	ClientKeyPassword string         `json:"clientKeyPassword"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

func (b *Broker) GetId() string              { return b.Id.String() }
func (b *Broker) GetTopics() map[string]byte { return map[string]byte{"$aws/#": 0} }
func (b *Broker) GetOptions() *mqtt.ClientOptions {
	cert, _ := tls.X509KeyPair(utils.StrToBytes(b.ClientCertificate), utils.StrToBytes(b.ClientKey))
	clientOptions := mqtt.NewClientOptions()
	clientOptions.AddBroker(fmt.Sprintf("%s:%d", b.Host, b.Port)).
		SetClientID(fmt.Sprintf("4suites-%s-%d", b.Id, time.Now().UnixNano())).
		SetProtocolVersion(4).
		SetTLSConfig(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})

	return clientOptions
}

func (b *Broker) GetEndpoint() string { return "/brokers" }
