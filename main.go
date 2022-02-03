package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/alex-held/dfctl/pkg/cli"
)

func main() {
	cmd := cli.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msgf("failed to execute command with args %#v", os.Args)
	}
}
