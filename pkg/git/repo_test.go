package git

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchLatestRelease(t *testing.T) {
	repo := NewRepoFromURL("gh:alex-held/dfctl")

	release, err := repo.FetchLatestRelease(http.DefaultClient)
	assert.NoError(t, err)

	assert.NotNil(t, release)
}
