package cli

import (
	"github.com/alex-held/dfctl/pkg/factory"
)

type CLI interface {
	Execute() (err error)
}

func New() CLI {
	f := factory.BuildFactory()
	root := NewRootCommand(f)
	return root
}
