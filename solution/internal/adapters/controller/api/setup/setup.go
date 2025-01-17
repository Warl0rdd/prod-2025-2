package setup

import (
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/spf13/viper"
	"solution/cmd/app"
	v1 "solution/internal/adapters/controller/api/v1"
	"solution/internal/adapters/controller/api/v1/b2b"
)

func Setup(app *app.App) {
	app.Fiber.Use(cors.New(cors.ConfigDefault))

	if viper.GetBool("settings.debug") {
		app.Fiber.Use(logger.New(logger.Config{TimeZone: viper.GetString("settings.timezone")}))
	}

	// Setup api v1 routes
	apiV1 := app.Fiber.Group("/api")

	pingHandler := v1.NewPingHandler()
	pingHandler.Setup(apiV1)

	businessHandler := b2b.NewBusinessHandler(app)
	businessHandler.Setup(apiV1)

	// Setup user routes
	//userHandler := v1.NewUserHandler(app)
	//userHandler.Setup(apiV1)
}
