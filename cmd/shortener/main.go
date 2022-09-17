package main

import (
	"github.com/tmitry/shorturl/internal/app"
	"github.com/tmitry/shorturl/internal/app/configs"
)

func main() {
	cfg := configs.NewConfig()

	app.StartServer(cfg)
}
