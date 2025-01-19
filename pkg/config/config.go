package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Server struct {
		Port string
	}
	JWT       JWTConfig
	MarketAPI struct {
		Key string
	}
	Environment string
}

type JWTConfig struct {
	Secret          string `env:"JWT_SECRET,required"`
	RefreshSecret   string `env:"JWT_REFRESH_SECRET,required"`
	ExpirationHours int    `env:"JWT_EXPIRATION_HOURS" envDefault:"24"`
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	config := &Config{}

	// Database configs
	config.Database.Host = viper.GetString("DB_HOST")
	config.Database.Port = viper.GetString("DB_PORT")
	config.Database.User = viper.GetString("DB_USER")
	config.Database.Password = viper.GetString("DB_PASSWORD")
	config.Database.Name = viper.GetString("DB_NAME")
	config.Database.SSLMode = viper.GetString("DB_SSL_MODE")

	// Server configs
	config.Server.Port = viper.GetString("API_PORT")

	// JWT configs
	config.JWT.Secret = viper.GetString("JWT_SECRET")

	// Market API configs
	config.MarketAPI.Key = viper.GetString("MARKET_API_KEY")

	config.Environment = viper.GetString("ENVIRONMENT")

	return config, nil
}
