package iostreams

import (
	"io"
	"os"
)

type IOStreams struct {
	In  io.ReadCloser
	Out io.WriteCloser
	Err io.WriteCloser
}

func Default() *IOStreams {
	return &IOStreams{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
}
