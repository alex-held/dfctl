package zsh

type InstallResult struct {
	Installed bool
	Err       error
}

type Installable interface {
	Id() string
	Install() (result InstallResult)
	IsInstalled() bool
	IsEnabled() bool
	Path() string
	GetKind() InstallableKind
	SetEnabled(enable bool) error
}
