package app

import (
	"github.com/gtgaleevtimur/gofermart/internal/config"
	"github.com/gtgaleevtimur/gofermart/internal/handler"
	"github.com/gtgaleevtimur/gofermart/internal/repository"
	serv "github.com/gtgaleevtimur/gofermart/internal/server"
	s "github.com/gtgaleevtimur/gofermart/internal/service"
	"github.com/rs/zerolog/log"
)

func Run() {
	log.Info().Msg("Start service `Gofermart`.")
	// Конфигурация приложения через считывание флагов и переменных окружения.
	conf := config.NewConfig()
	log.Info().Msg("The service will be started with the following configuration:")
	log.Info().Str("RUN_ADDRESS", conf.Address).
		Str("ACCRUAL_SYSTEM_ADDRESS", conf.AccrualAddress).
		Str("DATABASE_URI", conf.DatabaseDSN)
	// Инициализация хранилища сервиса.
	db, err := repository.NewDatabaseDSN(conf.DatabaseDSN)
	if err != nil {
		log.Info().Msg("Postgres init failed.")
		log.Fatal().Err(err)
	}
	// Инициализация сервиса.
	gophermart := s.NewService(db)
	// Инициализация роутера.
	hand := handler.NewHandler(gophermart)
	// Инициализация сервера.
	server := serv.NewServer(hand.R(), conf)

	// Запуск  сервера.
	go server.ListAndServ()

	blackbox := s.NewAccrl(db, conf.AccrualAddress)
	blackbox.Run()
}
