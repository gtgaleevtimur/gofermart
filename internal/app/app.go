package app

import (
	"github.com/gtgaleevtimur/gofermart/internal/config"
	r "github.com/gtgaleevtimur/gofermart/internal/repository"
	"github.com/rs/zerolog/log"
)

func Run() {
	conf := config.NewConfig()
	log.Debug().Str("RUN_ADDRESS", conf.Address).
		Str("DATABASE_URI", conf.DatabaseURI).
		Str("ACCRUAL_SYSTEM_ADDRESS", conf.AccrualSystemAddress).
		Msg("Receive config")
	repository, err := r.NewRepository(conf.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("Repository initialization failed")
	}

}
