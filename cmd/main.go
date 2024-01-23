package main

import (
	"github.com/poke-factory/cheri-berry/internal/config"
	"github.com/poke-factory/cheri-berry/internal/database"
	"github.com/poke-factory/cheri-berry/internal/routes"
	"github.com/poke-factory/cheri-berry/internal/storage"
	"log"
)

func main() {
	config.SetupConfig()
	database.ConnectDatabase()
	storage.SetupStorage()
	router := routes.SetupRouter()

	log.Printf("Starting server on %s", config.Cfg.ServerAddress)
	err := router.Run(config.Cfg.ServerAddress)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
