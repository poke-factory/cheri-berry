package storage

import (
	"github.com/poke-factory/cheri-berry/internal/config"
	"github.com/poke-factory/cheri-berry/pkg/soss"
	"log"
)

var Storage *soss.Soss

func SetupStorage() {
	oss, err := soss.New(soss.WithConfig(&config.Cfg.Config))
	if err != nil {
		log.Fatalf("Error loading storage: %v", err)
	}
	Storage = oss
}
