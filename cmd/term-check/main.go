// Package main provides the entry point for the GitHub application
package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/zendesk/term-check/internal/bot"
	"github.com/zendesk/term-check/internal/config"
)

var filepath = flag.String("config", "config.yaml", "Location of the configuration file.")
var debug = flag.Bool("debug", os.Getenv("LOG_LEVEL") == "debug", "sets log level to debug")

func main() {
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	flag.Parse()

	// Default logging level is info unless debug flag is present
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	c := config.New(*filepath)

	log.Info().Msg("Starting service...")
	bot.New(c.ForBot, c.ForClient, c.ForServer).Start()
}
