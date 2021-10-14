package git

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	git "github.com/go-git/go-git/v5"
	"github.com/google/go-cmp/cmp"
)

type mockGitSvc struct {
	cloneOpts   *git.CloneOptions
	fetchOpts   *git.FetchOptions
	plainOpened bool
}

func (g *mockGitSvc) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	g.cloneOpts = o
	return nil, nil
}

func (g *mockGitSvc) PlainOpen(path string) (*git.Repository, error) {
	g.plainOpened = true
	if strings.Contains(path, "Err") {
		return nil, errors.New("plain open goes boom")
	}

	if strings.HasSuffix(path, "myrepo3") {
		return &git.Repository{}, nil
	}
	return nil, nil
}

func (g *mockGitSvc) Fetch(r *git.Repository, o *git.FetchOptions) error {
	g.fetchOpts = o
	if r != nil {
		return git.NoErrAlreadyUpToDate
	}
	return nil
}

func (g *mockGitSvc) Worktree(r *git.Repository) (*git.Worktree, error) {
	return nil, nil
}

func (g *mockGitSvc) Checkout(w *git.Worktree, opts *git.CheckoutOptions) error {
	return nil
}

func newGitClient() (BasicClient, *mockGitSvc) {
	paths := []string{
		"myrepo/path/to/manifest.yaml",
		"myrepo2/path/to/manifest.yaml",
		"myrepo3/path/to/manifest.yaml",
		"plainopenErr/path/to/manifest.yaml",
	}
	mapFs := fstest.MapFS{}
	for _, path := range paths {
		mapFs[path] = &fstest.MapFile{
			Data: []byte("my bytes"),
		}
	}
	gitSvc := &mockGitSvc{}
	return BasicClient{
		auth: nil,
		mu:   &sync.Mutex{},
		git:  gitSvc,
		fs:   mapFs,
	}, gitSvc
}

type progressWriter struct{}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	return 0, nil
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
		{
			name:       "bubbles PlainOpen err",
			repository: "plainopenErr",
			commitHash: "123",
			path:       "path/to/manifest.yaml",
			errResult:  true,
			res:        "my bytes",
		},
	}

	pw := &progressWriter{}
	gitClient, gitSvc := newGitClient()
	WithProgressWriter(pw)(&gitClient)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			if gitSvc.plainOpened == false && gitSvc.cloneOpts.Progress != pw {
				t.Errorf("\ncloneOpts Progress not passed through: want: %v\n got: %v\n", pw, gitSvc.cloneOpts.Progress)
			}

			if gitSvc.fetchOpts.Progress != pw {
				t.Errorf("\nfetchOpts Progress not passed through: want: %v\n got: %v\n", pw, gitSvc.fetchOpts.Progress)
			}
		})
	}
}
