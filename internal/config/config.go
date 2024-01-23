package config

import (
	"github.com/poke-factory/cheri-berry/pkg/soss"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type AppConfig struct {
	ServerAddress string `yaml:"server_address" env:"SERVER_ADDRESS"`
	DSN           string `yaml:"dsn" env:"DSN"`
	JwtSecret     string `yaml:"jwt_secret" env:"JWT_SECRET"`
	soss.Config   `yaml:"storage"`
}

var Cfg = &AppConfig{}

func SetupConfig() {
	getConfigFromEnv := false
	if fromEnv, exists := os.LookupEnv("FROM_ENV"); exists {
		getConfigFromEnv = fromEnv == "true"
	}
	if getConfigFromEnv {
		loadConfigFromEnv(Cfg)
	} else {
		err := loadConfigFromYAML(Cfg, "config/config.yaml")
		if err != nil {
			log.Fatalf("Error loading config from YAML: %v", err)
		}
	}

}

func loadConfigFromEnv(cfg *AppConfig) {
	val := reflect.ValueOf(cfg).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		envVar := field.Tag.Get("env")
		if envVar == "" {
			envVar = strings.ToUpper(field.Name)
		}
		if envValue, exists := os.LookupEnv(envVar); exists {
			val.Field(i).SetString(envValue)
		}
	}
}

func loadConfigFromYAML(cfg *AppConfig, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	yamlFile, err := os.ReadFile(absPath)
	if err != nil {
		return nil // 文件不存在时不报错
	}

	return yaml.Unmarshal(yamlFile, cfg)
}
