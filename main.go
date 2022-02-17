package main

import (
	"github.com/alex-held/dfctl/pkg/cli"
	"github.com/alex-held/dfctl/pkg/errors"
)

func main() {
	app := cli.New()

	errors.Check(app.Execute())
}
