package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/git"
)

func TestManager_List(t *testing.T) {
	manager := NewManager(factory.Default)
	list, err := manager.list(true)
	assert.NoError(t, err)
	assert.NotEmpty(t, list)

}

func TestManager_Install(t *testing.T) {
	manager := NewManager(factory.Default)
	repo := git.NewGithubRepo("alex-held", "dfctl-hello-world")
	err := manager.Install(repo)
	assert.NoError(t, err)
}
