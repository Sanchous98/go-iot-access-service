package logger

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"bitbucket.org/4suites/iot-service-golang/pkg/infrastructure/unsafe"
	_ "embed"
	"github.com/Sanchous98/go-di"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func Factory(di.Container) logger.Logger {
	var loggerConfig zap.Config
	if err := yaml.Unmarshal(unsafe.Must(io.ReadAll(unsafe.Must(os.Open("config/logger.yaml")))), &loggerConfig); err != nil {
		panic(err)
	}
	return unsafe.Must(loggerConfig.Build()).WithOptions().Sugar()
}
