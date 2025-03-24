package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type DB struct {
	DBConnString string `yaml:"db_conn_string" env:"DB_CONN_STRING"`
	PoolMinConns int32  `yaml:"pool_min_conns" env:"POOL_MIN_CONNS"`
}

type Config struct {
	LogLevel    string        `yaml:"log_level" env:"LOG_LEVEL"`
	AppAddress  string        `yaml:"app_address" env:"APP_ADDRESS"`
	ReadTimeout time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT"`
	DB          DB            `yaml:"db"`
}

func MustLoad(configPath string) *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return &cfg
}
