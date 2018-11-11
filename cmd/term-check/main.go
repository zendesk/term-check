package main

import (
	"strconv"

	"github.com/ragurney/term-check/internal/bot"
	"github.com/ragurney/term-check/pkg/lib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = ""

	id, err := strconv.Atoi(lib.Env("APP_ID", ""))
	if err != nil {
		log.Fatal().Err(err)
	}
	pk := lib.Env("PRIVATE_KEY_PATH", "")
	ws := lib.Env("WEBHOOK_SECRET_KEY", "")

	log.Info().Msg("Starting service...")

	bot.New(
		bot.WithAppID(id),
		bot.WithPrivateKeyPath(pk),
		bot.WithWebhookSecretKey(ws),
	).Start()
}
