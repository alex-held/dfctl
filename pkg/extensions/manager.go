package extensions

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/cli/cli/pkg/findsh"
	"github.com/cli/safeexec"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/alex-held/dfctl/pkg/iostreams"

	"github.com/alex-held/dfctl/pkg/dfpath"
	"github.com/alex-held/dfctl/pkg/factory"
	"github.com/alex-held/dfctl/pkg/git"
)

type Manager struct {
	dataDir    string
	lookPath   func(file string) (string, error)
	findSh     func() (string, error)
	newCommand func(name string, arg ...string) *exec.Cmd
	io         *iostreams.IOStreams
	fs         afero.Fs
	client     *http.Client
}

func (m *Manager) List(includeMetadata bool) (extensions []Extension) {
	exts, err := m.list(includeMetadata)
	if err != nil {
		return extensions
	}

	for _, ext := range exts {
		var extension = ext
		extensions = append(extensions, &extension)
	}

	return extensions
}

func (m *Manager) list(includeMetadata bool) ([]extension, error) {
	dir := m.dataDir
	entries, err := afero.ReadDir(m.fs, dir)
	if err != nil {
		return nil, err
	}

	var results []extension
	for _, f := range entries {
		if !strings.HasPrefix(f.Name(), "dfctl-") {
			continue
		}
		var ext extension
		var err error
		if f.IsDir() {
			ext, err = m.parseExtensionDir(f)
			if err != nil {
				return nil, err
			}
			results = append(results, ext)
		} else {
			ext, err = m.parseExtensionFile(f)
			if err != nil {
				return nil, err
			}
			results = append(results, ext)
		}
	}

	if includeMetadata {
		m.populateLatestVersions(results)
	}

	return results, nil
}

func (m *Manager) Install(repo git.Repository) error {
	isBin, err := isBinExtension(m.client, repo)
	if err != nil {
		return fmt.Errorf("could not check for binary extension: %w", err)
	}
	if isBin {
		return m.installBin(repo)
	}

	return fmt.Errorf("repo currently unsupported")

	// hs, err := hasScript(m.client, repo)
	// if err != nil {
	// 	return err
	// }
	// if !hs {
	// 	return errors.New("extension is not installable: missing executable")
	// }
	//
	// protocol, _ := m.config.GetOrDefault(repo.RepoHost(), "git_protocol")
	// return m.installGit(ghrepo.FormatRemoteURL(repo, protocol), m.io.Out, m.io.Err)
}

func isBinExtension(client *http.Client, repo git.Repository) (isBin bool, err error) {
	var r *git.Release
	r, err = repo.FetchLatestRelease(client)
	if err != nil {
		return false, err
	}

	for _, a := range r.Assets {
		dists := possibleDists()
		for _, d := range dists {
			suffix := d
			if strings.HasPrefix(d, "windows") {
				suffix += ".exe"
			}
			if strings.HasSuffix(a.Name, suffix) {
				isBin = true
				break
			}
		}
	}

	return isBin, err
}

func possibleDists() []string {
	return []string{
		"darwin-amd64",
		"darwin-arm64",
		"freebsd-386",
		"freebsd-amd64",
		"freebsd-arm",
		"freebsd-arm64",
		"illumos-amd64",
		"linux-386",
		"linux-amd64",
		"linux-arm",
		"linux-arm64",
		"linux-mips",
		"linux-mips64",
		"linux-mips64le",
		"linux-mipsle",
		"linux-ppc64",
		"linux-ppc64le",
		"linux-riscv64",
		"linux-s390x",
		"netbsd-386",
		"netbsd-amd64",
		"netbsd-arm",
		"netbsd-arm64",
		"openbsd-386",
		"openbsd-amd64",
		"openbsd-arm",
		"openbsd-arm64",
		"openbsd-mips64",
		"plan9-386",
		"plan9-amd64",
		"plan9-arm",
		"solaris-amd64",
	}
}

func (m *Manager) InstallLocal(dir string) error {
	// TODO implement me
	panic("implement me")
}

func (m *Manager) Upgrade(name string, force bool) error {
	// TODO implement me
	panic("implement me")
}

func (m *Manager) Remove(name string) error {
	// TODO implement me
	panic("implement me")
}

func (m *Manager) Dispatch(args []string, stdin io.Reader, stdout, stderr io.Writer) (bool, error) {
	if len(args) == 0 {
		return false, fmt.Errorf("too few arguments in list")
	}

	var exe string
	extName := args[0]
	forwardArgs := args[1:]

	log.Debug().Str("extension", extName).Str("args", fmt.Sprintf("%v", forwardArgs)).Msg("running extension")

	exts, _ := m.list(true)
	var ext Extension
	for _, e := range exts {
		log.Debug().Msgf("extension %s", e.Name())
		if e.Name() == extName {
			ext = &e
			exe = ext.Path()
			break
		}
	}
	if exe == "" {
		return false, nil
	}

	var externalCmd *exec.Cmd
	if ext.IsBinary() {
		externalCmd = m.newCommand(exe, forwardArgs...)
	}

	externalCmd.Stdin = stdin
	externalCmd.Stdout = stdout
	externalCmd.Stderr = stderr
	return true, externalCmd.Run()
}

func (m *Manager) Create(name string, tmplType ExtTemplateType) error {
	// TODO implement me
	panic("implement me")
}

func (m *Manager) parseExtensionFile(fi fs.FileInfo) (extension, error) {
	ext := extension{isLocal: true}
	id := m.dataDir
	exePath := filepath.Join(id, fi.Name(), fi.Name())
	if !isSymlink(fi.Mode()) {
		// if this is a regular file, its contents is the local directory of the extension
		p, err := readPathFromFile(filepath.Join(id, fi.Name()))
		if err != nil {
			return ext, err
		}
		exePath = filepath.Join(p, fi.Name())
	}
	ext.path = exePath
	return ext, nil
}

// reads the product of makeSymlink on Windows
func readPathFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b := make([]byte, 1024)
	n, err := f.Read(b)
	return strings.TrimSpace(string(b[:n])), err
}

func isSymlink(m os.FileMode) bool {
	return m&os.ModeSymlink != 0
}

func (m *Manager) parseExtensionDir(fi fs.FileInfo) (extension, error) {
	id := m.dataDir
	if _, err := os.Stat(filepath.Join(id, fi.Name(), manifestName)); err == nil {
		return m.parseBinaryExtensionDir(fi)
	}

	return m.parseGitExtensionDir(fi)
}

type binManifest struct {
	Owner string
	Name  string
	Host  string
	Tag   string
	Path  string
}

func (m *Manager) parseBinaryExtensionDir(fi fs.FileInfo) (extension, error) {
	id := m.dataDir
	exePath := filepath.Join(id, fi.Name(), fi.Name())

	ext := extension{path: exePath, kind: BinaryKind}
	manifestPath := filepath.Join(id, fi.Name(), manifestName)
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return ext, fmt.Errorf("could not open %s for reading: %w", manifestPath, err)
	}
	var bm binManifest
	err = yaml.Unmarshal(manifest, &bm)
	if err != nil {
		return ext, fmt.Errorf("could not parse %s: %w", manifestPath, err)
	}
	repo := git.NewRepoWithHost(bm.Host, bm.Owner, bm.Name)
	remoteURL := repo.URI()
	ext.url = remoteURL
	ext.currentVersion = bm.Tag
	return ext, nil
}

func (m *Manager) parseGitExtensionDir(fi fs.FileInfo) (extension, error) {
	id := m.dataDir
	exePath := filepath.Join(id, fi.Name(), fi.Name())
	remoteUrl := m.getRemoteUrl(fi.Name())
	currentVersion := m.getCurrentVersion(fi.Name())
	return extension{
		path:           exePath,
		url:            remoteUrl,
		isLocal:        false,
		currentVersion: currentVersion,
		kind:           GitKind,
	}, nil
}

// getRemoteUrl determines the remote URL for non-local git extensions.
func (m *Manager) getRemoteUrl(extension string) string {
	gitExe, err := m.lookPath("git")
	if err != nil {
		return ""
	}
	dir := m.dataDir
	gitDir := "--git-dir=" + filepath.Join(dir, extension, ".git")
	cmd := m.newCommand(gitExe, gitDir, "config", "remote.origin.url")
	url, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(url))
}

// getCurrentVersion determines the current version for non-local git extensions.
func (m *Manager) getCurrentVersion(extension string) string {
	gitExe, err := m.lookPath("git")
	if err != nil {
		return ""
	}
	dir := m.dataDir
	gitDir := "--git-dir=" + filepath.Join(dir, extension, ".git")
	cmd := m.newCommand(gitExe, gitDir, "rev-parse", "HEAD")
	localSha, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(localSha))
}

func (m *Manager) populateLatestVersions(exts []extension) {
	size := len(exts)
	type result struct {
		index   int
		version string
	}
	ch := make(chan result, size)
	var wg sync.WaitGroup
	wg.Add(size)
	for idx, ext := range exts {
		go func(i int, e extension) {
			defer wg.Done()
			version, _ := m.getLatestVersion(e)
			ch <- result{index: i, version: version}
		}(idx, ext)
	}
	wg.Wait()
	close(ch)
	for r := range ch {
		ext := &exts[r.index]
		ext.latestVersion = r.version
	}
}

func (m *Manager) getLatestVersion(ext extension) (string, error) {
	if ext.isLocal {
		localExtensionUpgradeError := fmt.Errorf("local extension is not upgradable")
		return "", localExtensionUpgradeError
	}
	if ext.IsBinary() {
		repo := git.NewRepoFromURL(ext.url)

		r, err := repo.FetchLatestRelease(http.DefaultClient)
		if err != nil {
			return "", err
		}
		return r.Tag, nil
	} else {
		gitExe, err := m.lookPath("git")
		if err != nil {
			return "", err
		}
		extDir := filepath.Dir(ext.path)
		gitDir := "--git-dir=" + filepath.Join(extDir, ".git")
		cmd := m.newCommand(gitExe, gitDir, "ls-remote", "origin", "HEAD")
		lsRemote, err := cmd.Output()
		if err != nil {
			return "", err
		}
		remoteSha := bytes.SplitN(lsRemote, []byte("\t"), 2)[0]
		return string(remoteSha), nil
	}
}

func (m *Manager) installBin(repo git.Repository) error {
	var r *git.Release
	r, err := repo.FetchLatestRelease(m.client)
	if err != nil {
		return err
	}

	platform, ext := m.platform()
	var asset *git.Asset
	for _, a := range r.Assets {
		if strings.HasSuffix(a.Name, platform+ext) {
			asset = &a
			break
		}
	}

	if asset == nil {
		return fmt.Errorf(
			"%[1]s unsupported for %[2]s. Open an issue: `gh issue create -R %[3]s/%[1]s -t'Support %[2]s'`",
			repo.GetName(), platform, repo.GetUser())
	}

	name := repo.GetName()
	targetDir := filepath.Join(m.dataDir, name)
	// TODO clean this up if function errs?
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	binPath := filepath.Join(targetDir, name)
	binPath += ext

	err = asset.Download(m.client, binPath)
	if err != nil {
		return fmt.Errorf("failed to download asset %s: %w", asset.Name, err)
	}

	manifest := binManifest{
		Name:  name,
		Owner: repo.GetUser(),
		Host:  repo.GetHost(),
		Path:  binPath,
		Tag:   r.Tag,
	}

	bs, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}

	manifestPath := filepath.Join(targetDir, manifestName)

	f, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open manifest for writing: %w", err)
	}
	defer f.Close()

	_, err = f.Write(bs)
	if err != nil {
		return fmt.Errorf("failed write manifest file: %w", err)
	}

	return nil
}

//
// func hasScript(httpClient *http.Client, repo git.Repository) (hs bool, err error) {
// 	path := fmt.Sprintf("repos/%s/%s/contents/%s", repo.GetUser(), repo.RepoName(), repo.RepoName())
// 	url := ghinstance.RESTPrefix(repo.RepoHost()) + path
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return
// 	}
//
// 	resp, err := httpClient.Do(req)
// 	if err != nil {
// 		return
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode == 404 {
// 		return
// 	}
//
// 	if resp.StatusCode > 299 {
// 		err = fmt.Errorf("unable to check if repo has script. code: %v, status: %v", resp.StatusCode, resp.Status)
// 		return
// 	}
//
// 	hs = true
// 	return
// }

func (m *Manager) platform() (platform string, ext string) {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH), ""
}

func NewManager(factory *factory.Factory) ExtensionManager {
	return &Manager{
		dataDir:    dfpath.Extensions(),
		lookPath:   safeexec.LookPath,
		findSh:     findsh.Find,
		newCommand: exec.Command,
		io:         factory.Streams,
		fs:         factory.Fs,
		client:     http.DefaultClient,
	}
}
