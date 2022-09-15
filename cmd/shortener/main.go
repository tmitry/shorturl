package main

import (
	"github.com/tmitry/shorturl/internal/app"
	"github.com/tmitry/shorturl/internal/app/configs"
)

func main() {
	configs.InitConfigs()
	app.StartServer()
}
