package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port      string `mapstructure:"PORT"`
	DBUrl     string `mapstructure:"DB_URL"`
	RedisUrl  string `mapstructure:"REDIS_URL"`
	JWTSecret string `mapstructure:"JWT_SECRET"`
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Try reading .env file if it exists, otherwise use OS env
	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal("Cannot decode config:", err)
	}

	// Fallbacks
	if cfg.Port == "" {
		if os.Getenv("PORT") != "" {
			cfg.Port = os.Getenv("PORT")
		} else {
			cfg.Port = "8080"
		}
	}
	if cfg.DBUrl == "" {
		cfg.DBUrl = os.Getenv("DB_URL")
	}
	if cfg.RedisUrl == "" {
		cfg.RedisUrl = os.Getenv("REDIS_URL")
	}
	if cfg.JWTSecret == "" {
		if os.Getenv("JWT_SECRET") != "" {
			cfg.JWTSecret = os.Getenv("JWT_SECRET")
		} else {
			cfg.JWTSecret = "development_secret_do_not_use_in_production"
		}
	}

	return &cfg
}
