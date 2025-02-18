package main

import (
	"prod/cmd/app"
	"prod/internal/adapters/config"
	"prod/internal/adapters/controller/api/setup"
)

func main() {
	appConfig := config.Configure()
	mainApp := app.New(appConfig)

	setup.Setup(mainApp)
	mainApp.Start()
}
