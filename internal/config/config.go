package config

import (
	"flag"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ObjectStorage struct {
	Address         string `yaml:"address"           env:"MINIO_ADDRESS"`
	EndPoint        string `yaml:"endpoint"          env:"MINIO_ENDPOINT"         env-required:"true"`
	AccessKeyID     string `yaml:"access_key_id"     env:"MINIO_ACCESS_KEY_ID"     env-required:"true"`
	SecretAccessKey string `yaml:"secret_access_key" env:"MINIO_SECRET_ACCESS_KEY" env-required:"true"`
	BucketName      string `yaml:"bucket_name"       env:"MINIO_BUCKET_NAME"       env-required:"true"`
}

type DB struct {
	URL     string `yaml:"url"      env:"POSTGRES_URL"      env-required:"true"`
	PoolMax int    `yaml:"pool_max" env:"POSTGRES_POOL_MAX" env-default:"10"`
}

type Server struct {
	RestPort int `yaml:"rest_port" env:"REST_PORT" env-default:"8080"`
	GRPCPort int `yaml:"grpc_port" env:"GRPC_PORT" env-default:"50053"`
}

type Token struct {
	TTL        time.Duration `yaml:"ttl"         env:"TOKEN_TTL"         env-default:"1h"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env:"REFRESH_TOKEN_TTL" env-default:"24h"`
}

type Kafka struct {
	Brokers string `yaml:"brokers" env:"KAFKA_BROKERS" env-required:"true"`
	Topic   string `yaml:"topic"   env:"KAFKA_TOPIC"   env-required:"true"`
	GroupID string `yaml:"group_id" env:"KAFKA_GROUP_ID" env-default:"drip-mate"`
}

type Config struct {
	Server        Server        `yaml:"server"`
	DB            DB            `yaml:"db"`
	Token         Token         `yaml:"token"`
	ObjectStorage ObjectStorage `yaml:"object_storage"`
	Kafka         Kafka         `yaml:"kafka"`
}

func MustLoadConfig() *Config {
	path := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig(*path, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
