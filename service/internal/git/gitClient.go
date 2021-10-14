package git

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

// Option is a function for configuring the BasicClient
type Option func(*BasicClient)

func WithProgressWriter(w io.Writer) Option {
	return func(c *BasicClient) {
		c.pw = w
	}
}

// BasicClient connects to git using ssh
type BasicClient struct {
	auth    transport.AuthMethod
	mu      *sync.Mutex
	git     gitSvc
	fs      fs.FS
	baseDir string // base directory to run git operations from
	pw      io.Writer
}

// NewSSHBasicClient creates a new ssh based git client
func NewSSHBasicClient(sshPemFile string, opts ...Option) (BasicClient, error) {
	auth, err := ssh.NewPublicKeysFromFile("git", sshPemFile, "")
	if err != nil {
		return BasicClient{}, err
	}

	return newBasicClient(auth, opts...), nil
}

// NewHTTPSBasicClient creates a new https based git client
func NewHTTPSBasicClient(user, pass string, opts ...Option) (BasicClient, error) {
	auth := &http.BasicAuth{
		Username: user,
		Password: pass,
	}

	return newBasicClient(auth, opts...), nil
}

func newBasicClient(auth transport.AuthMethod, opts ...Option) BasicClient {
	cl := BasicClient{
		auth:    auth,
		mu:      &sync.Mutex{},
		git:     gitSvcImpl{},
		fs:      os.DirFS(os.TempDir()),
		baseDir: os.TempDir(),
		pw:      ioutil.Discard,
	}

	for _, o := range opts {
		o(&cl)
	}

	return cl
}

func (g BasicClient) GetManifestFile(repository, commitHash, path string) ([]byte, error) {
	// filePath should only be used for git calls. direct fs calls should use repository directly
	repPath := strings.ReplaceAll(repository, "/", "")
	filePath := filepath.Join(g.baseDir, repPath)

	// Locking here since we need to make sure nobody else is using the repo at the same time to ensure the right sha is checked out
	// TODO: use a lock per repository instead of a single global lock
	g.mu.Lock()
	defer g.mu.Unlock()

	var repo *git.Repository

	if _, err := fs.Stat(g.fs, repPath); os.IsNotExist(err) {
		// TODO: use context version and make depth configurable
		repo, err = g.git.PlainClone(filePath, false, &git.CloneOptions{
			URL:      repository,
			Auth:     g.auth,
			Progress: g.pw,
		})
		if err != nil {
			return []byte{}, err
		}
	} else {
		repo, err = g.git.PlainOpen(filePath)
		if err != nil {
			return []byte{}, err
		}
		err = g.git.Fetch(repo, &git.FetchOptions{
			Progress: g.pw,
			Auth:     g.auth,
		})
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

	pathToManifest := filepath.Join(repPath, path)
	fileStat, err := fs.Stat(g.fs, pathToManifest)
	if err != nil {
		return []byte{}, err
	}

	if fileStat.IsDir() {
		return []byte{}, fmt.Errorf("path provided is not a file '%s'", path)
	}

	return fs.ReadFile(g.fs, pathToManifest)
}
