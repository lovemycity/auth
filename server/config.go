package server

import (
	"github.com/caarlos0/env"
	"github.com/pkg/errors"
)

// Config is the server configuration struct
type Config struct {
	Port          string `env:"PORT" envDefault:"3000"`
	MongoDSN      string `env:"MONGO_DSN" envDefault:"mongodb://localhost:27017/auth"`
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisNetwork  string `env:"REDIS_NETWORK" envDefault:"tcp"`
	SessionName   string `env:"SESSION_NAME" envDefault:"sid"`
	SessionSecret string `env:"SESSION_SECRET" envDefault:"secret"`
	SessionDomain string `env:"SESSION_DOMAIN" envDefault:"lovemycity.local"`
}

func getConfig() (*Config, error) {
	c := new(Config)
	if err := env.Parse(c); err != nil {
		return nil, errors.Wrap(err, "failed to parse config")
	}
	return c, nil
}
