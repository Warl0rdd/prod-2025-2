package app

import (
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"os"
	"prod/internal/adapters/config"
	"prod/internal/adapters/controller/api/validator"
	"prod/internal/adapters/logger"
)

// App is a struct that contains the fiber app, database connection, listen port, validator, logging boolean etc.
type App struct {
	Fiber     *fiber.App
	DB        *gorm.DB
	Redis     *redis.Client
	Validator *validator.Validator
}

// New is a function that creates a new app struct
func New(config *config.Config) *App {
	fiberApp := fiber.New(fiber.Config{
		// Global custom error handler
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(validator.GlobalErrorHandlerResp{
				Success: false,
				Message: err.Error(),
			})
		},
	},
	)

	return &App{
		Fiber:     fiberApp,
		DB:        config.Database,
		Redis:     config.Redis,
		Validator: validator.New(),
	}
}

// Start is a function that starts the app
func (a *App) Start() {
	if viper.GetBool("settings.listen-tls") {
		if err := a.Fiber.Listen(
			":"+os.Getenv("SERVER_PORT"),
			fiber.ListenConfig{
				CertFile:    viper.GetString("service.backend.certificate.cert-file"),
				CertKeyFile: viper.GetString("service.backend.certificate.key-file"),
			}); err != nil {
			logger.Log.Panicf("failed to start listen (with tls): %v", err)
		}
	} else {
		logger.Log.Debugf("port: %s", viper.GetString("service.backend.port"))
		if err := a.Fiber.Listen(":" + os.Getenv("SERVER_PORT")); err != nil {
			logger.Log.Panicf("failed to start listen (no tls): %v", err)
		}
	}
}
