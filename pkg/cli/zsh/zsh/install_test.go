package zsh

import (
	"testing"
)

func TestGetInstallablesByNames(t *testing.T) {
	sut := (interface{})(nil)

	err := runInstallCommand()
	if err != nil {
		return
	}
}
