package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	GRPCServerPort int `env:"GRPC_SERVER_PORT" env-description:"grpc server port" env-default:"50053"`

	Address         string `env:"ADDRESS"`
	EndPoint        string `env:"ENDPOINT" env-default:"s3.storage.selcloud.ru"`
	AccessKeyId     string `env:"ACCESS_KEY_ID"`
	SecretAccessKey string `env:"SECRET_ACCESS_KEY"`
	BucketName      string `env:"BUCKET_NAME"`
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
