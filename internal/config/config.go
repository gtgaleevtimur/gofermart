package config

import (
	"flag"
	"github.com/caarlos0/env"
)

// Config -структура конфигурационного файла приложения.
type Config struct {
	Address        string `env:"RUN_ADDRESS"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseDSN    string `env:"DATABASE_URI"`
}

// NewConfig - конструктор конфигурационного файла.
func NewConfig(options ...Option) *Config {
	conf := Config{}

	//если в аргументах получили Options, то применяем их к Config.
	for _, opt := range options {
		opt(&conf)
	}
	return &conf
}

// Option - функция применяемая к Config для его заполнения.
type Option func(*Config)

// WithParseEnv - парсит из окружения/флагов, изменяет Config.
func WithParseEnv() Option {
	return func(c *Config) {
		env.Parse(c)
		c.ParseFlags()
	}
}

// ParseFlags - парсит флаги.
func (c *Config) ParseFlags() {
	flag.StringVar(&c.Address, "a", ":8080", "SERVER_ADDRESS")
	flag.StringVar(&c.AccrualAddress, "r", c.AccrualAddress, "ACCRUAL_SYSTEM_ADDRESS")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "DATABASE_URI")
	flag.Parse()
}
