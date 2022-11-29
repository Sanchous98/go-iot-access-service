package api

import (
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

	handlers []Handler `inject:"api.handler"`
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
