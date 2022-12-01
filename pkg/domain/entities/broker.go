package entities

import (
	"github.com/google/uuid"
)

type Broker struct {
	Id uuid.UUID `json:"id"`
	//UserId            uuid.UUID     `json:"userId"`
	//Name              string         `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
	//Claimed           bool           `json:"claimed"`
	//Enabled           bool           `json:"enabled"`
	//Metadata          map[string]any `json:"metadata"`
	//CaCertificate     string         `json:"caCertificate"`
	ClientCertificate string `json:"clientCertificate"`
	ClientKey         string `json:"clientKey"`
	//ClientKeyPassword string         `json:"clientKeyPassword"`
	//CreatedAt         time.Time      `json:"createdAt"`
	//UpdatedAt         time.Time      `json:"updatedAt"`
}

func (b *Broker) GetId() uuid.UUID           { return b.Id }
func (b *Broker) GetTopics() map[string]byte { return map[string]byte{"$aws/#": 0} }
func (*Broker) GetResource() string          { return "brokers" }
