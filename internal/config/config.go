package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Logger  LoggerConf  `toml:"logger"`
	RedisDB RedisDBConf `toml:"redis"`
	Limiter LimiterConf `toml:"limiter"`
	Server  ServerConf  `toml:"server"`
}

type LoggerConf struct {
	Level  string   `toml:"level"`
	Out    []string `toml:"out"`
	OutErr []string `toml:"out_err"`
}

type RedisDBConf struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type LimiterConf struct {
	N int64 `toml:"n"`
	M int64 `toml:"m"`
	K int64 `toml:"k"`
}

type ServerConf struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

func NewConfig(configFile string) *Config {
	var cfg Config
	err := cleanenv.ReadConfig(configFile, &cfg)
	if err != nil {
		log.Fatalf("cannot read config, err: %v", err)
	}
	return &cfg
}
