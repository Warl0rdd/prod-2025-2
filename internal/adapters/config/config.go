package config

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"os"
	postgresRepo "prod/internal/adapters/database/postgres"
	"prod/internal/adapters/logger"
	"time"
)

type Config struct {
	Database *gorm.DB
	Redis    *redis.Client
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("failed to read config: %v", err)
	}
}

func Configure() *Config {
	initConfig()

	logger.New(
		viper.GetBool("settings.debug"),
		viper.GetString("settings.timezone"),
	)
	logger.Log.Debugf("Debug mode: %t", viper.GetBool("settings.debug"))

	// Initialize database
	logger.Log.Info("Initializing database...")
	logger.Log.Debug("Configuring database logger")
	var gormConfig *gorm.Config
	if viper.GetBool("settings.debug") {
		newLogger := gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormLogger.Config{
				SlowThreshold: time.Second,
				LogLevel:      gormLogger.Info,
				Colorful:      true,
			},
		)
		gormConfig = &gorm.Config{
			Logger: newLogger,
		}
	}

	logger.Log.Debug("Configuring postgres connection string")
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s TimeZone=%s",
		//viper.GetString("service.database.user"),
		//viper.GetString("service.database.password"),
		//viper.GetString("service.database.name"),
		//viper.GetString("service.database.host"),
		//viper.GetString("service.database.port"),
		os.Getenv("POSTGRES_USERNAME"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DATABASE"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		viper.GetString("service.database.ssl-mode"),
		viper.GetString("settings.timezone"),
	)

	logger.Log.Debugf("dsn: %s", dsn)
	logger.Log.Debug("Connecting to postgres...")
	database, errConnect := gorm.Open(postgres.Open(dsn), gormConfig)
	if errConnect != nil {
		logger.Log.Panicf("Failed to connect to postgres: %v", errConnect)
	} else {
		logger.Log.Info("Connected to postgres")
	}

	logger.Log.Info("Running migrations...")
	if errMigrate := database.AutoMigrate(postgresRepo.Migrations...); errMigrate != nil {
		logger.Log.Panicf("Failed to run migrations: %v", errMigrate)
	}

	logger.Log.Info("Database initialized")

	logger.Log.Info("Initializing redis...")
	redisAddress := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})
	logger.Log.Info("Redis initialized")

	return &Config{
		Database: database,
		Redis:    redisClient,
	}
}
