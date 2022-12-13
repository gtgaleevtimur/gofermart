package config

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

// Config -структура конфигурационного файла приложения.
type Config struct {
	Address        string `env:"RUN_ADDRESS"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseDSN    string `env:"DATABASE_URI"`
}

// NewConfig - конструктор конфигурационного файла.
func NewConfig() *Config {
	conf := Config{}
	flag.StringVar(&conf.Address, "a", ":8080", "SERVER_ADDRESS")
	flag.StringVar(&conf.AccrualAddress, "r", "http://localhost:8081", "ACCRUAL_SYSTEM_ADDRESS")
	flag.StringVar(&conf.DatabaseDSN, "d", "", "DATABASE_URI")
	flag.Parse()

	err := env.Parse(conf)
	if err != nil {
		log.Fatalln(err)
	}
	return &conf
}
