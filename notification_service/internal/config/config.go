package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	RestServerPort int `env:"REST_SERVER_PORT" env-description:"rest server port" env-default:"8081"`
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
