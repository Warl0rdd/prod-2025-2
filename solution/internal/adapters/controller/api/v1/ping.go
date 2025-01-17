package v1

import (
	"github.com/gofiber/fiber/v3"
)

type PingHandler struct{}

func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

func (h PingHandler) ping(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).SendString("GOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOL")
}

func (h PingHandler) Setup(router fiber.Router) {
	router.Get("/ping", h.ping)
}
