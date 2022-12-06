package services

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"sync"
)

type ClientsPool struct {
	mu      sync.RWMutex
	clients map[string]mqtt.Client
}

func (c *ClientsPool) All() []mqtt.Client {
	c.mu.RLock()
	all := make([]mqtt.Client, 0, len(c.clients))

	for _, client := range c.clients {
		all = append(all, client)
	}

	c.mu.RUnlock()

	return all
}

func (c *ClientsPool) CreateClient(options *mqtt.ClientOptions) mqtt.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.clients[options.ClientID]; !ok {
		if c.clients == nil {
			c.clients = make(map[string]mqtt.Client, 1)
		}

		c.clients[options.ClientID] = mqtt.NewClient(options)
	}

	return c.clients[options.ClientID]
}

func (c *ClientsPool) GetClient(clientId string) mqtt.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if client, ok := c.clients[clientId]; ok {
		return client
	}

	return nil
}

func (c *ClientsPool) DeleteClient(clientId string) {
	c.mu.Lock()
	c.clients[clientId].Disconnect(250)
	delete(c.clients, clientId)
	c.mu.Unlock()
}

func (c *ClientsPool) Purge() {
	for clientId := range c.clients {
		c.DeleteClient(clientId)
	}
}
