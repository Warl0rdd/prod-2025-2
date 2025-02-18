package setup

import (
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/spf13/viper"
	"prod/cmd/app"
	v1 "prod/internal/adapters/controller/api/v1"
	"prod/internal/adapters/controller/api/v1/b2b"
	"prod/internal/adapters/controller/api/v1/b2c"
	"prod/internal/adapters/controller/api/v1/middlewares"
)

func Setup(app *app.App) {
	app.Fiber.Use(cors.New(cors.ConfigDefault))

	if viper.GetBool("settings.debug") {
		app.Fiber.Use(logger.New(logger.Config{TimeZone: viper.GetString("settings.timezone")}))
	}

	// Setup api v1 routes
	apiV1 := app.Fiber.Group("/api")

	middlewareHandler := middlewares.NewMiddlewareHandler(app)

	pingHandler := v1.NewPingHandler()
	pingHandler.Setup(apiV1)

	businessHandler := b2b.NewBusinessHandler(app)
	businessHandler.Setup(apiV1)

	promoHandler := b2b.NewPromoHandler(app)
	promoHandler.Setup(apiV1, middlewareHandler.IsAuthenticated())

	// Setup user routes
	userAuthHandler := b2c.NewUserHandler(app)
	userAuthHandler.Setup(apiV1, middlewareHandler.IsAuthenticated())

	userPromoHandler := b2c.NewUserPromoHandler(app)
	userPromoHandler.Setup(apiV1, middlewareHandler.IsAuthenticated())

	userActionsHandler := b2c.NewActionsHandler(app)
	userActionsHandler.Setup(apiV1, middlewareHandler.IsAuthenticated())
}
