package api

import "github.com/gofiber/fiber/v2"

type ServerApi struct {
	*fiber.App
	host string `env:"API_HOST"`
	port string `env:"API_PORT"`
}

func (s *ServerApi) Constructor() {
	s.App = fiber.New()
}

func (s *ServerApi) Launch() {
	_ = s.Listen(s.host + ":" + s.port)
}
