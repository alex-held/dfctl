package errors

import (
	"github.com/rs/zerolog/log"
)

func Check(err error) {
	if err != nil {
		log.Fatal().Err(err).Msgf("error check failed: %w", err)
	}
}
