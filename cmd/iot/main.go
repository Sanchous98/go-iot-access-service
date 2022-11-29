package main

import (
	"bitbucket.org/4suites/iot-service-golang/cmd/internal"
	"bitbucket.org/4suites/iot-service-golang/iot/api"
	"bitbucket.org/4suites/iot-service-golang/pkg/http"
	"context"
	"github.com/Sanchous98/go-di"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if err := recover(); err != nil {
			log.Printf("%v", err)
			cancel()
		}
	}()

	app := di.Application(ctx)
	app.AddEntryPoint(internal.Bootstrap)
	app.Set(new(http.ServerApi))
	app.Set(new(api.IotApiHandler), "api.handler")
}
