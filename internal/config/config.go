package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	RestServerPort int    `env:"REST_SERVER_PORT" env-required:"true" env-description:"rest server port"`
	URL            string `json:"url" env:"URL" env-required:"true" env-description:"db connection string"`
	PoolMax        int    `json:"poolMax" env:"POOL_MAX" env-required:"true" env-description:"db pool max"`
}

func MustLoadConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	cfg := Config{}
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
