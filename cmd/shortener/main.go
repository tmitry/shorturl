package main

import (
	"github.com/tmitry/shorturl/internal/app"
	"github.com/tmitry/shorturl/internal/app/config"
)

func main() {
	config.InitConfigs()
	app.StartServer()
}
