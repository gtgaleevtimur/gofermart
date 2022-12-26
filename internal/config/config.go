package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	Address              string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

// NewConfig - функция конструктор конфига с настройками окружения.
func NewConfig() *Config {
	c := &Config{}
	flag.StringVar(&c.Address, "a", ":8080", "RUN_ADDRESS")
	flag.StringVar(&c.DatabaseURI, "d", "", "DATABASE_URI")
	flag.StringVar(&c.AccrualSystemAddress, "r", "http://localhost:8081", "ACCRUAL_SYSTEM_ADDRESS")
	flag.Parse()
	err := env.Parse(c)
	if err != nil {
		log.Fatalln(err)
	}
	return c
}
