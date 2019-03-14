// Package main provides the entry point for the GitHub application
package main

import (
	"flag"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/zendesk/term-check/internal/bot"
	"github.com/zendesk/term-check/internal/config"
)

var filepath = flag.String("config", "config.yaml", "Location of the configuration file.")

func main() {
	zerolog.TimeFieldFormat = ""
	flag.Parse()

	c := config.New(*filepath)

	log.Info().Msg("Starting service...")
	bot.New(c.ForBot, c.ForClient, c.ForServer).Start()
}
