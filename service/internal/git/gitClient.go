package git

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// Client allows for retrieving data from git repo
type Client interface {
	GetManifestFile(repository, commitHash, path string) ([]byte, error)
}

type gitSvc interface {
	PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error)
	PlainOpen(path string) (*git.Repository, error)
	Fetch(r *git.Repository, o *git.FetchOptions) error
	Worktree(r *git.Repository) (*git.Worktree, error)
	Checkout(w *git.Worktree, opts *git.CheckoutOptions) error
}

type gitSvcImpl struct{}

func (g gitSvcImpl) PlainClone(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, o)
}

func (g gitSvcImpl) PlainOpen(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

func (g gitSvcImpl) Fetch(r *git.Repository, o *git.FetchOptions) error {
	return r.Fetch(o)
}

func (g gitSvcImpl) Worktree(r *git.Repository) (*git.Worktree, error) {
	return r.Worktree()
}

func (g gitSvcImpl) Checkout(w *git.Worktree, opts *git.CheckoutOptions) error {
	return w.Checkout(opts)
}

type osSvc interface {
	fs.StatFS
}

type osSvcImpl struct{}

func (o osSvcImpl) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (o osSvcImpl) Open(name string) (fs.File, error) {
	return os.Open(name)
}

// BasicClient connects to git using ssh
type BasicClient struct {
	auth *ssh.PublicKeys
	mu   *sync.Mutex
	git  gitSvc
	os   osSvc
}

// NewBasicClient creates a new ssh based git client
func NewBasicClient(sshPemFile string) (BasicClient, error) {
	auth, err := ssh.NewPublicKeysFromFile("git", sshPemFile, "")
	if err != nil {
		return BasicClient{}, err
	}

	return BasicClient{
		auth: auth,
		mu:   &sync.Mutex{},
		git:  gitSvcImpl{},
		os:   osSvcImpl{},
	}, nil
}

func (g BasicClient) GetManifestFile(repository, commitHash, path string) ([]byte, error) {
	filePath := filepath.Join(os.TempDir(), repository)

	// Locking here since we need to make sure nobody else is using the repo at the same time to ensure the right sha is checked out
	// TODO: use a lock per repository instead of a single global lock
	g.mu.Lock()
	defer g.mu.Unlock()

	var repo *git.Repository

	if _, err := g.os.Stat(filePath); os.IsNotExist(err) {
		// TODO: use context version and make depth configurable
		repo, err = g.git.PlainClone(filePath, false, &git.CloneOptions{
			URL:  repository,
			Auth: g.auth,
		})
		if err != nil {
			return []byte{}, err
		}
	} else {
		repo, err = g.git.PlainOpen(filePath)
		if err != nil {
			return []byte{}, err
		}
		err = g.git.Fetch(repo, &git.FetchOptions{})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return []byte{}, err
		}
	}

	w, err := g.git.Worktree(repo)
	if err != nil {
		return []byte{}, err
	}

	err = g.git.Checkout(w, &git.CheckoutOptions{
		Hash: plumbing.NewHash(commitHash),
	})
	if err != nil {
		return []byte{}, err
	}

	pathToManifest := filepath.Join(filePath, path)
	fileStat, err := g.os.Stat(pathToManifest)
	if err != nil {
		return []byte{}, err
	}

	if fileStat.IsDir() {
		return []byte{}, fmt.Errorf("path provided is not a file '%s'", path)
	}

	return fs.ReadFile(g.os, pathToManifest)
}
