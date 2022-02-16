package extensions

import (
	"path/filepath"
	"strings"
)

const manifestName = "manifest.yaml"

type ExtensionKind int

const (
	GitKind ExtensionKind = iota
	BinaryKind
)

type extension struct {
	path           string
	kind           ExtensionKind
	url            string
	currentVersion string
	isLocal        bool
	latestVersion  string
}

func (e *extension) Name() string {
	return strings.TrimPrefix(filepath.Base(e.path), "dfctl-")
}

func (e *extension) Path() string {
	return e.path
}

func (e *extension) URL() string {
	return e.url
}

func (e *extension) IsLocal() bool {
	return e.isLocal
}

func (e *extension) UpdateAvailable() bool {
	if e.isLocal ||
		e.currentVersion == "" ||
		e.latestVersion == "" ||
		e.currentVersion == e.latestVersion {
		return false
	}
	return true
}

func (e *extension) IsBinary() bool {
	return e.kind == BinaryKind
}
