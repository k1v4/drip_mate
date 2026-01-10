package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/k1v4/drip_mate/user_service/pkg/DataBase/postgres"
)

type Config struct {
	postgres.DBConfig

	GRPCServerPort  int           `env:"GRPC_SERVER_PORT" env-description:"grpc server port" env-default:"50051"`
	RestServerPort  int           `env:"REST_SERVER_PORT" env-description:"rest server port" env-default:"8080"`
	TokenTTL        time.Duration `env:"TOKEN_TTL" env-default:"1h"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" env-default:"24h"`
}

func MustLoadConfig() *Config {
	err := godotenv.Load(".env") // Явно указываем путь
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
