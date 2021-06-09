package git

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// GitClient allows for retrieving data from git repo
type GitClient interface {
	GetManifestFile(repository, commitHash, path string) ([]byte, error)
}

// BasicGitClient connects to git using ssh
type BasicGitClient struct {
	auth *ssh.PublicKeys
	mu   *sync.Mutex
}

// NewBasicGitClient creates a new ssh based git client
func NewBasicGitClient(sshPemFile string) (BasicGitClient, error) {
	auth, err := ssh.NewPublicKeysFromFile("git", sshPemFile, "")
	if err != nil {
		return BasicGitClient{}, err
	}

	return BasicGitClient{
		auth: auth,
		mu:   &sync.Mutex{},
	}, nil
}

func (g BasicGitClient) GetManifestFile(repository, commitHash, path string) ([]byte, error) {
	filePath := filepath.Join(os.TempDir(), repository)

	// Locking here since we need to make sure nobody else is using the repo at the same time to ensure the right sha is checked out
	// TODO: use a lock per repository instead of a single global lock
	g.mu.Lock()
	defer g.mu.Unlock()

	var repo *git.Repository

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// TODO: use context version and make depth configurable
		repo, err = git.PlainClone(filePath, false, &git.CloneOptions{
			URL:  repository,
			Auth: g.auth,
		})
		if err != nil {
			return []byte{}, err
		}
	} else {
		repo, err = git.PlainOpen(filePath)
		if err != nil {
			return []byte{}, err
		}
		err = repo.Fetch(&git.FetchOptions{})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return []byte{}, err
		}
	}

	w, err := repo.Worktree()
	if err != nil {
		return []byte{}, err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitHash),
	})
	if err != nil {
		return []byte{}, err
	}

	pathToManifest := filepath.Join(filePath, path)
	fileStat, err := os.Stat(pathToManifest)
	if err != nil {
		return []byte{}, err
	}

	if fileStat.IsDir() {
		return []byte{}, fmt.Errorf("path provided is not a file '%s'", path)
	}

	file, err := os.Open(pathToManifest)
	if err != nil {
		return []byte{}, err
	}

	fileContents := make([]byte, fileStat.Size())
	_, err = file.Read(fileContents)

	return fileContents, err
}
