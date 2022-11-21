package api

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

type ServerApi struct {
	*fiber.App
	host string `env:"API_HOST"`
	port string `env:"API_PORT"`
}

func (s *ServerApi) Constructor() {
	s.App = fiber.New()
}

func (s *ServerApi) Launch(context context.Context) {
	s.App.Use(func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context)
		return ctx.Next()
	})
	_ = s.Listen(s.host + ":" + s.port)
}
