package logger

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type mqttErrorWrapper struct {
	log logger.Logger `inject:""`
}

func NewMqttErrorWrapper(log logger.Logger) mqtt.Logger {
	return &mqttErrorWrapper{log: log}
}

func (m *mqttErrorWrapper) Println(v ...any) {
	m.log.Errorln(v...)
}

func (m *mqttErrorWrapper) Printf(format string, v ...any) {
	m.log.Errorf(format, v)
}

type mqttWarnWrapper struct {
	log logger.Logger `inject:""`
}

func NewMqttWarnWrapper(log logger.Logger) mqtt.Logger {
	return &mqttWarnWrapper{log: log}
}

func (m *mqttWarnWrapper) Println(v ...any) {
	m.log.Warnln(v...)
}

func (m *mqttWarnWrapper) Printf(format string, v ...any) {
	m.log.Warnf(format, v)
}

type mqttDebugWrapper struct {
	log logger.Logger `inject:""`
}

func NewMqttDebugWrapper(log logger.Logger) mqtt.Logger {
	return &mqttDebugWrapper{log: log}
}

func (m *mqttDebugWrapper) Println(v ...any) {
	m.log.Debugln(v...)
}

func (m *mqttDebugWrapper) Printf(format string, v ...any) {
	m.log.Debugf(format, v)
}

type mqttCriticalWrapper struct {
	log logger.Logger `inject:""`
}

func NewMqttCriticalWrapper(log logger.Logger) mqtt.Logger {
	return &mqttCriticalWrapper{log: log}
}

func (m *mqttCriticalWrapper) Println(v ...any) {
	m.log.Fatalln(v...)
}

func (m *mqttCriticalWrapper) Printf(format string, v ...any) {
	m.log.Fatalf(format, v)
}
