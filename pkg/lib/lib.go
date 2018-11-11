package lib

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

// Env looks env value for passed in key, logging and failing if not set
func Env(name string, fallback string) string {
	zerolog.TimeFieldFormat = ""

	v, ok := os.LookupEnv(name)
	if !ok {
		if fallback != "" {
			return fallback
		}
		log.Fatal().Msgf("Environment variable is not set: %s", name)
	}
	return v
}

// Contains -
func Contains(set map[string]struct{}, item string) bool {
	_, ok := set[item]
	return ok
}

// Unique -
func Unique(slice []string) (res []string) {
	set := make(map[string]struct{})

	for _, s := range slice {
		if Contains(set, s) {
			continue
		}
		res = append(res, s)
		set[s] = struct{}{}
	}

	return res
}
