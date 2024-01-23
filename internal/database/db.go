package database

import (
	"github.com/poke-factory/cheri-berry/internal/config"
	"github.com/poke-factory/cheri-berry/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(postgres.Open(config.Cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!", err)
	}

	err = database.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database!", err)
		return
	}

	DB = database
}
