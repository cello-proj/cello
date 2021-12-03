package git

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
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
	pcErr       error
	poErr       error
	fetchErr    error
	wtErr       error
	coErr       error
}

func (g *mockGitSvc) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	g.cloneOpts = o

	if g.pcErr != nil {
		return nil, g.pcErr
	}

	return nil, nil
}

func (g *mockGitSvc) PlainOpen(path string) (*git.Repository, error) {
	fmt.Println(path)
	g.plainOpened = true

	if g.poErr != nil {
		return nil, g.poErr
	}

	if strings.HasSuffix(path, "myrepo3") {
		return &git.Repository{}, nil
	}

	return nil, nil
}

func (g *mockGitSvc) Fetch(r *git.Repository, o *git.FetchOptions) error {
	g.fetchOpts = o
	if g.fetchErr != nil {
		return g.fetchErr
	}

	if r != nil {
		return git.NoErrAlreadyUpToDate
	}

	return nil
}

func (g *mockGitSvc) Worktree(r *git.Repository) (*git.Worktree, error) {
	if g.wtErr != nil {
		return nil, g.wtErr
	}

	return nil, nil
}

func (g *mockGitSvc) Checkout(w *git.Worktree, opts *git.CheckoutOptions) error {
	if g.coErr != nil {
		return g.coErr
	}

	return nil
}

func newGitClient() (BasicClient, *mockGitSvc) {
	paths := []string{
		"myrepo/path/to/manifest.yaml",
		"myrepo2/path/to/manifest.yaml",
		"myrepo3/path/to/manifest.yaml",
	}
	mapFs := fstest.MapFS{}
	mapFs["aDir/aPath"] = &fstest.MapFile{
		Mode: os.ModeDir,
	}
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

func TestGetManifestErrors(t *testing.T) {
	tests := []struct {
		name string
		repo string
		path string

		pc     error
		po     error
		fetch  error
		wt     error
		co     error
		errStr string
	}{
		{
			name: "bubbles PlainClone error",
			repo: "plainclone",
			pc:   errors.New("PlainClone err"),
		},
		{
			name: "bubbles PlainOpen error",
			po:   errors.New("PlainOpen err"),
		},
		{
			name:  "bubbles Fetch error",
			fetch: errors.New("Fetch err"),
		},
		{
			name: "bubbles WorkTree error",
			wt:   errors.New("WorkTree err"),
		},
		{
			name: "bubbles Checkout error",
			co:   errors.New("Checkout err"),
		},
		{
			name:   "rejects when path is a dir",
			repo:   "aDir",
			path:   "aPath",
			errStr: "path provided is not a file",
		},
	}

	cl, svc := newGitClient()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc.pcErr = tt.pc
			svc.poErr = tt.po
			svc.fetchErr = tt.fetch
			svc.wtErr = tt.wt
			svc.coErr = tt.co

			repo := defaultString(tt.repo, "myrepo3")
			path := defaultString(tt.path, "path/to/manifest.yaml")
			_, err := cl.GetManifestFile(repo, "123", path)

			for _, want := range []error{tt.pc, tt.po, tt.fetch, tt.wt, tt.co} {
				if want != nil && !errors.Is(err, want) {
					t.Errorf("wanted: %+v got: %+v", want, err)
				}
			}

			if tt.errStr != "" && !strings.Contains(err.Error(), tt.errStr) {
				t.Errorf("wanted: %+v got: %+v\n", tt.errStr, err)
			}
		})
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

			if !gitSvc.plainOpened && gitSvc.cloneOpts.Progress != pw {
				t.Errorf("\ncloneOpts Progress not passed through: want: %v\n got: %v\n", pw, gitSvc.cloneOpts.Progress)
			}

			if gitSvc.fetchOpts.Progress != pw {
				t.Errorf("\nfetchOpts Progress not passed through: want: %v\n got: %v\n", pw, gitSvc.fetchOpts.Progress)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	t.Run("NewSSHBasicClient creates client with ssh auth with valid PEM", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "tmpssh*.pem")
		assertNoErr(t, err)
		defer tmp.Close()

		pk, _ := rsa.GenerateKey(rand.Reader, 2048)
		asn := x509.MarshalPKCS1PrivateKey(pk)
		err = pem.Encode(tmp, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: asn})
		assertNoErr(t, err)

		cl, err := NewSSHBasicClient(tmp.Name())
		assertNoErr(t, err)
		want := "ssh-public-keys"

		if cl.auth.Name() != want {
			t.Errorf("client auth, want: %s got: %s\n", want, cl.auth.Name())
		}
	})

	t.Run("NewSSHBasicClient rejects with invalid PEM", func(t *testing.T) {
		_, err := NewSSHBasicClient("fakefile.pem")
		if err == nil {
			t.Error("expected error, received nil")
		}
	})

	t.Run("NewHTTPSBasicClient creates client with http auth", func(t *testing.T) {
		cl, err := NewHTTPSBasicClient("user", "pass")
		assertNoErr(t, err)

		want := "http-basic-auth"

		if cl.auth.Name() != want {
			t.Errorf("client auth, want: %s got: %s\n", want, cl.auth.Name())
		}
	})
	t.Run("NewHTTPSBasicClient passes opts", func(t *testing.T) {
		pw := &progressWriter{}
		cl, err := NewHTTPSBasicClient("user", "pass", WithProgressWriter(pw))
		assertNoErr(t, err)

		if cl.pw != pw {
			t.Errorf("want: %+v got: %+v\n", pw, cl.pw)
		}
	})
}

func assertNoErr(t *testing.T, err error) {
	if err == nil {
		return
	}
	t.Errorf("unexpected err: %+v\n", err)
}

func defaultString(in, def string) string {
	if in == "" {
		return def
	}

	return in
}
