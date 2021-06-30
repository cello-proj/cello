package git

import (
	"io/fs"
	"strings"
	"sync"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
)

type mockGitSvc struct{}

func (g mockGitSvc) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return nil, nil
}

func (g mockGitSvc) PlainOpen(path string) (*git.Repository, error) {
	if strings.HasSuffix(path, "myrepo3") {
		return &git.Repository{}, nil
	}
	return nil, nil
}

func (g mockGitSvc) Fetch(r *git.Repository, o *git.FetchOptions) error {
	if r != nil {
		return git.NoErrAlreadyUpToDate
	}
	return nil
}

func (g mockGitSvc) Worktree(r *git.Repository) (*git.Worktree, error) {
	return nil, nil
}

func (g mockGitSvc) Checkout(w *git.Worktree, opts *git.CheckoutOptions) error {
	return nil
}

type mockOsSvc struct{}

func (o mockOsSvc) Stat(name string) (fs.FileInfo, error) {
	if strings.HasSuffix(name, "myrepo2") {
		return nil, fs.ErrNotExist
	}

	return nil, nil
}

func (o mockOsSvc) Open(name string) (fs.File, error) {
	return nil, nil
}

func (o mockOsSvc) IsDir(filestat fs.FileInfo) bool {
	return false
}

func (o mockOsSvc) Size(filestat fs.FileInfo) int64 {
	return 8
}

func (o mockOsSvc) Read(file fs.File, b []byte) (n int, err error) {
	copy(b, "my bytes")
	return 0, nil
}

func newGitClient() BasicClient {
	return BasicClient{
		auth: nil,
		mu:   &sync.Mutex{},
		git:  mockGitSvc{},
		os:   mockOsSvc{},
	}
}

func TestGetManifestFile(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		commitHash string
		path       string
		errResult  bool
		res        string
	}{
		{
			name:       "get manifest exists on fs success",
			repository: "myrepo",
			commitHash: "123",
			path:       "path/to/manifest.yaml",
			errResult:  false,
			res:        "my bytes",
		},
		{
			name:       "get manifest new clone success",
			repository: "myrepo2",
			commitHash: "123",
			path:       "path/to/manifest.yaml",
			errResult:  false,
			res:        "my bytes",
		},
		{
			name:       "get manifest fetch already updated",
			repository: "myrepo3",
			commitHash: "123",
			path:       "path/to/manifest.yaml",
			errResult:  false,
			res:        "my bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitClient := newGitClient()
			res, err := gitClient.GetManifestFile(tt.repository, tt.commitHash, tt.path)
			if err != nil {
				if !tt.errResult {
					t.Errorf("\ndid not expect error, got: %v", err)
				}
			} else {
				if tt.errResult {
					t.Errorf("\nexpected error")
				}
				if !cmp.Equal(string(res), tt.res) {
					t.Errorf("\nwant: %v\n got: %v", tt.res, string(res))
				}
			}
		})
	}
}
