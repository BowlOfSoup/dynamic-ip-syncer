package control

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"time"
)

func InitLogger() string {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	logLevel := os.Getenv("LOGLEVEL")
	logLevelAsInt, _ := strconv.ParseInt(logLevel, 10, 8)
	zerolog.SetGlobalLevel(zerolog.Level(logLevelAsInt))

	return zerolog.GlobalLevel().String()
}
