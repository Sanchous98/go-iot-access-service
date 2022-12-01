package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"crypto/tls"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2/utils"
	"log"
	"sync"
	"time"
)

var optionsPool sync.Map

func GetClientOptions(b *entities.Broker) *mqtt.ClientOptions {
	if clientOptions, ok := optionsPool.Load(b.Id.String()); ok {
		return clientOptions.(*mqtt.ClientOptions)
	}

	cert, _ := tls.X509KeyPair(utils.UnsafeBytes(b.ClientCertificate), utils.UnsafeBytes(b.ClientKey))
	clientOptions := mqtt.NewClientOptions()
	clientOptions.AddBroker(fmt.Sprintf("%s:%d", b.Host, b.Port)).
		SetClientID(fmt.Sprintf("4suites-broker-%s", b.Id.String())).
		SetProtocolVersion(4).
		SetTLSConfig(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}).
		SetCleanSession(false).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectTimeout(5 * time.Second).
		SetConnectRetryInterval(5 * time.Second).
		SetOrderMatters(false).
		SetReconnectingHandler(func(mqtt.Client, *mqtt.ClientOptions) {
			log.Printf("Broker %s reconnecting\n", b.Id)
		}).
		SetOnConnectHandler(func(mqtt.Client) {
			log.Printf("Broker %s connected\n", b.Id)
		})

	return clientOptions
}
