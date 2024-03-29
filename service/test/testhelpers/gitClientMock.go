// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package testhelpers

import (
	"github.com/cello-proj/cello/service/internal/git"
	"sync"
)

// Ensure, that GitClientMock does implement git.Client.
// If this is not the case, regenerate this file with moq.
var _ git.Client = &GitClientMock{}

// GitClientMock is a mock implementation of git.Client.
//
// 	func TestSomethingThatUsesClient(t *testing.T) {
//
// 		// make and configure a mocked git.Client
// 		mockedClient := &GitClientMock{
// 			GetManifestFileFunc: func(repository string, commitHash string, path string) ([]byte, error) {
// 				panic("mock out the GetManifestFile method")
// 			},
// 		}
//
// 		// use mockedClient in code that requires git.Client
// 		// and then make assertions.
//
// 	}
type GitClientMock struct {
	// GetManifestFileFunc mocks the GetManifestFile method.
	GetManifestFileFunc func(repository string, commitHash string, path string) ([]byte, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetManifestFile holds details about calls to the GetManifestFile method.
		GetManifestFile []struct {
			// Repository is the repository argument value.
			Repository string
			// CommitHash is the commitHash argument value.
			CommitHash string
			// Path is the path argument value.
			Path string
		}
	}
	lockGetManifestFile sync.RWMutex
}

// GetManifestFile calls GetManifestFileFunc.
func (mock *GitClientMock) GetManifestFile(repository string, commitHash string, path string) ([]byte, error) {
	if mock.GetManifestFileFunc == nil {
		panic("GitClientMock.GetManifestFileFunc: method is nil but Client.GetManifestFile was just called")
	}
	callInfo := struct {
		Repository string
		CommitHash string
		Path       string
	}{
		Repository: repository,
		CommitHash: commitHash,
		Path:       path,
	}
	mock.lockGetManifestFile.Lock()
	mock.calls.GetManifestFile = append(mock.calls.GetManifestFile, callInfo)
	mock.lockGetManifestFile.Unlock()
	return mock.GetManifestFileFunc(repository, commitHash, path)
}

// GetManifestFileCalls gets all the calls that were made to GetManifestFile.
// Check the length with:
//     len(mockedClient.GetManifestFileCalls())
func (mock *GitClientMock) GetManifestFileCalls() []struct {
	Repository string
	CommitHash string
	Path       string
} {
	var calls []struct {
		Repository string
		CommitHash string
		Path       string
	}
	mock.lockGetManifestFile.RLock()
	calls = mock.calls.GetManifestFile
	mock.lockGetManifestFile.RUnlock()
	return calls
}
