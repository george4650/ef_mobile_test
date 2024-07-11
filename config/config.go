package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Http
		Grpc
		Ldap
		Oracle
		Postgres
		Samba
		Auth
		Links
	}

	Http struct {
		Port int
	}

	Grpc struct {
		Port int
	}

	Ldap struct {
		Servers []string
		Domains string
	}

	Postgres struct {
		Host     string
		Port     int
		User     string
		Password string
		Dbname   string
		Sslmode  string
	}

	Oracle struct {
		Host     string
		Port     int
		Service  string
		User     string
		Password string
	}

	Samba struct {
		Host     string
		Port     int
		Domain   string
		User     string
		Password string
	}

	Auth struct {
		Key string
	}

	Links struct {
		Audio string
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("config file read error: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return cfg, nil
}
