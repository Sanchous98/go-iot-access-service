package http

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"context"
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	RegisterRoutes(app *fiber.App)
}

type ServerApi struct {
	*fiber.App
	host string `env:"API_HOST"`
	port string `env:"API_PORT"`

	handlers []Handler     `inject:"api.handler"`
	log      logger.Logger `inject:""`
}

func (s *ServerApi) Constructor() {
	s.App = fiber.New()

	for _, handler := range s.handlers {
		handler.RegisterRoutes(s.App)
	}
}

func (s *ServerApi) Launch(context context.Context) {
	s.App.Use(func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context)
		return ctx.Next()
	})
	if err := s.Listen(s.host + ":" + s.port); err != nil {
		panic(err)
	}
}

func (s *ServerApi) Shutdown(context.Context) {
	if err := s.App.Shutdown(); err != nil {
		s.log.Errorln(err)
	} else {
		s.log.Infoln("Fiber has successfully shut down")
	}
}
