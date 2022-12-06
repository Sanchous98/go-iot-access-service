package application

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/entities"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"crypto/tls"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2/utils"
	"sync"
	"time"
)

var optionsPool sync.Map

func GetClientOptions(log logger.Logger, b *entities.Broker) *mqtt.ClientOptions {
	if clientOptions, ok := optionsPool.Load(b.Id.String()); ok {
		return clientOptions.(*mqtt.ClientOptions)
	}

	cert, _ := tls.X509KeyPair(utils.UnsafeBytes(b.ClientCertificate), utils.UnsafeBytes(b.ClientKey))

	return mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("%s:%d", b.Host, b.Port)).
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
			log.Warnf("Broker %s reconnecting\n", b.Id)
		}).
		SetOnConnectHandler(func(mqtt.Client) {
			log.Warnf("Broker %s connected\n", b.Id)
		})
}
