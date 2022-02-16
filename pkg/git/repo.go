package git

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

type RepositoryKind int

type Repository interface {
	fmt.Stringer
	URI() string
	GetUser() string
	GetHost() string
	GetName() string

	FetchLatestRelease(client *http.Client) (release *Release, err error)
}

func NewGithubRepo(user, repo string) Repository {
	return NewRepoWithHost("github.com", user, repo)
}

type repository struct {
	Host string
	User string
	Name string
}

// FetchLatestRelease finds the latest published release for a repository.
func (r *repository) FetchLatestRelease(httpClient *http.Client) (release *Release, err error) {
	url := r.APIURI("releases/latest")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("unable to fetch latest release. status: %v; %v", resp.StatusCode, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	release = &Release{}
	err = json.Unmarshal(b, release)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func (r *repository) GetUser() string {
	return r.User
}

func (r *repository) GetHost() string {
	return r.Host
}

func (r *repository) GetName() string {
	return r.Name
}

func (r *repository) Clone(path string) (err error) {
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL: r.URI(),
	})
	return err
}

func (r *repository) APIURI(paths ...string) string {
	path := ""
	if len(paths) > 0 {
		path = "/" + strings.Join(paths, "/")
	}
	return fmt.Sprintf("https://api.github.com/repos/%s/%s%s", r.User, r.Name, path)
}

func (r *repository) String() string {
	return r.URI()
}

func (r *repository) URI() string {
	return fmt.Sprintf("https://%s/%s/%s", r.Host, r.User, r.Name)
}

func NewRepoFromURL(urlString string) Repository {
	if strings.HasPrefix(urlString, "gh:") {
		urlString = strings.TrimPrefix(urlString, "gh:")
		return NewGithubRepo(strings.Split(urlString, "/")[0], strings.Split(urlString, "/")[1])
	}

	if strings.HasPrefix(urlString, "git:") {
		urlString = strings.TrimPrefix(urlString, "git:")
	}

	url, err := url.Parse(urlString)
	if err != nil {
		panic(err)
	}
	host := url.Host
	user := strings.Split(url.Path, "/")[0]
	name := strings.Split(url.Path, "/")[1]
	return NewRepoWithHost(host, user, name)
}

func NewRepoWithHost(host, user, repo string) Repository {
	return &repository{Host: host, User: user, Name: repo}
}

type Release struct {
	Url    string  `json:"url"`
	Id     int     `json:"id"`
	Tag    string  `json:"tag_name"`
	TarUrl string  `json:"tarball_url"`
	ZipUrl string  `json:"zipball_url"`
	Assets []Asset `json:"assets"`
}

type Asset struct {
	Url  string `json:"url"`
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (a *Asset) Download(client *http.Client, destPath string) error {
	req, err := http.NewRequest("GET", a.Url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("unable to download asset %s; code: %v; status: %v", a.Name, resp.StatusCode, resp.Status)
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
