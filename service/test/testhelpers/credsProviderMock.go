// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package testhelpers

import (
	"github.com/cello-proj/cello/internal/responses"
	"github.com/cello-proj/cello/internal/types"
	"github.com/cello-proj/cello/service/internal/credentials"
	"sync"
)

// Ensure, that CredsProviderMock does implement credentials.Provider.
// If this is not the case, regenerate this file with moq.
var _ credentials.Provider = &CredsProviderMock{}

// CredsProviderMock is a mock implementation of credentials.Provider.
//
// 	func TestSomethingThatUsesProvider(t *testing.T) {
//
// 		// make and configure a mocked credentials.Provider
// 		mockedProvider := &CredsProviderMock{
// 			CreateProjectFunc: func(s string) (types.Token, error) {
// 				panic("mock out the CreateProject method")
// 			},
// 			CreateTargetFunc: func(s string, target types.Target) error {
// 				panic("mock out the CreateTarget method")
// 			},
// 			CreateTokenFunc: func(s string) (types.Token, error) {
// 				panic("mock out the CreateToken method")
// 			},
// 			DeleteProjectFunc: func(s string) error {
// 				panic("mock out the DeleteProject method")
// 			},
// 			DeleteProjectTokenFunc: func(s1 string, s2 string) error {
// 				panic("mock out the DeleteProjectToken method")
// 			},
// 			DeleteTargetFunc: func(s1 string, s2 string) error {
// 				panic("mock out the DeleteTarget method")
// 			},
// 			GetProjectFunc: func(s string) (responses.GetProject, error) {
// 				panic("mock out the GetProject method")
// 			},
// 			GetProjectTokenFunc: func(s1 string, s2 string) (types.ProjectToken, error) {
// 				panic("mock out the GetProjectToken method")
// 			},
// 			GetTargetFunc: func(s1 string, s2 string) (types.Target, error) {
// 				panic("mock out the GetTarget method")
// 			},
// 			GetTokenFunc: func() (string, error) {
// 				panic("mock out the GetToken method")
// 			},
// 			ListTargetsFunc: func(s string) ([]string, error) {
// 				panic("mock out the ListTargets method")
// 			},
// 			ProjectExistsFunc: func(s string) (bool, error) {
// 				panic("mock out the ProjectExists method")
// 			},
// 			TargetExistsFunc: func(s1 string, s2 string) (bool, error) {
// 				panic("mock out the TargetExists method")
// 			},
// 			UpdateTargetFunc: func(s string, target types.Target) error {
// 				panic("mock out the UpdateTarget method")
// 			},
// 		}
//
// 		// use mockedProvider in code that requires credentials.Provider
// 		// and then make assertions.
//
// 	}
type CredsProviderMock struct {
	// CreateProjectFunc mocks the CreateProject method.
	CreateProjectFunc func(s string) (types.Token, error)

	// CreateTargetFunc mocks the CreateTarget method.
	CreateTargetFunc func(s string, target types.Target) error

	// CreateTokenFunc mocks the CreateToken method.
	CreateTokenFunc func(s string) (types.Token, error)

	// DeleteProjectFunc mocks the DeleteProject method.
	DeleteProjectFunc func(s string) error

	// DeleteProjectTokenFunc mocks the DeleteProjectToken method.
	DeleteProjectTokenFunc func(s1 string, s2 string) error

	// DeleteTargetFunc mocks the DeleteTarget method.
	DeleteTargetFunc func(s1 string, s2 string) error

	// GetProjectFunc mocks the GetProject method.
	GetProjectFunc func(s string) (responses.GetProject, error)

	// GetProjectTokenFunc mocks the GetProjectToken method.
	GetProjectTokenFunc func(s1 string, s2 string) (types.ProjectToken, error)

	// GetTargetFunc mocks the GetTarget method.
	GetTargetFunc func(s1 string, s2 string) (types.Target, error)

	// GetTokenFunc mocks the GetToken method.
	GetTokenFunc func() (string, error)

	// ListTargetsFunc mocks the ListTargets method.
	ListTargetsFunc func(s string) ([]string, error)

	// ProjectExistsFunc mocks the ProjectExists method.
	ProjectExistsFunc func(s string) (bool, error)

	// TargetExistsFunc mocks the TargetExists method.
	TargetExistsFunc func(s1 string, s2 string) (bool, error)

	// UpdateTargetFunc mocks the UpdateTarget method.
	UpdateTargetFunc func(s string, target types.Target) error

	// calls tracks calls to the methods.
	calls struct {
		// CreateProject holds details about calls to the CreateProject method.
		CreateProject []struct {
			// S is the s argument value.
			S string
		}
		// CreateTarget holds details about calls to the CreateTarget method.
		CreateTarget []struct {
			// S is the s argument value.
			S string
			// Target is the target argument value.
			Target types.Target
		}
		// CreateToken holds details about calls to the CreateToken method.
		CreateToken []struct {
			// S is the s argument value.
			S string
		}
		// DeleteProject holds details about calls to the DeleteProject method.
		DeleteProject []struct {
			// S is the s argument value.
			S string
		}
		// DeleteProjectToken holds details about calls to the DeleteProjectToken method.
		DeleteProjectToken []struct {
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
		}
		// DeleteTarget holds details about calls to the DeleteTarget method.
		DeleteTarget []struct {
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
		}
		// GetProject holds details about calls to the GetProject method.
		GetProject []struct {
			// S is the s argument value.
			S string
		}
		// GetProjectToken holds details about calls to the GetProjectToken method.
		GetProjectToken []struct {
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
		}
		// GetTarget holds details about calls to the GetTarget method.
		GetTarget []struct {
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
		}
		// GetToken holds details about calls to the GetToken method.
		GetToken []struct {
		}
		// ListTargets holds details about calls to the ListTargets method.
		ListTargets []struct {
			// S is the s argument value.
			S string
		}
		// ProjectExists holds details about calls to the ProjectExists method.
		ProjectExists []struct {
			// S is the s argument value.
			S string
		}
		// TargetExists holds details about calls to the TargetExists method.
		TargetExists []struct {
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
		}
		// UpdateTarget holds details about calls to the UpdateTarget method.
		UpdateTarget []struct {
			// S is the s argument value.
			S string
			// Target is the target argument value.
			Target types.Target
		}
	}
	lockCreateProject      sync.RWMutex
	lockCreateTarget       sync.RWMutex
	lockCreateToken        sync.RWMutex
	lockDeleteProject      sync.RWMutex
	lockDeleteProjectToken sync.RWMutex
	lockDeleteTarget       sync.RWMutex
	lockGetProject         sync.RWMutex
	lockGetProjectToken    sync.RWMutex
	lockGetTarget          sync.RWMutex
	lockGetToken           sync.RWMutex
	lockListTargets        sync.RWMutex
	lockProjectExists      sync.RWMutex
	lockTargetExists       sync.RWMutex
	lockUpdateTarget       sync.RWMutex
}

// CreateProject calls CreateProjectFunc.
func (mock *CredsProviderMock) CreateProject(s string) (types.Token, error) {
	if mock.CreateProjectFunc == nil {
		panic("CredsProviderMock.CreateProjectFunc: method is nil but Provider.CreateProject was just called")
	}
	callInfo := struct {
		S string
	}{
		S: s,
	}
	mock.lockCreateProject.Lock()
	mock.calls.CreateProject = append(mock.calls.CreateProject, callInfo)
	mock.lockCreateProject.Unlock()
	return mock.CreateProjectFunc(s)
}

// CreateProjectCalls gets all the calls that were made to CreateProject.
// Check the length with:
//     len(mockedProvider.CreateProjectCalls())
func (mock *CredsProviderMock) CreateProjectCalls() []struct {
	S string
} {
	var calls []struct {
		S string
	}
	mock.lockCreateProject.RLock()
	calls = mock.calls.CreateProject
	mock.lockCreateProject.RUnlock()
	return calls
}

// CreateTarget calls CreateTargetFunc.
func (mock *CredsProviderMock) CreateTarget(s string, target types.Target) error {
	if mock.CreateTargetFunc == nil {
		panic("CredsProviderMock.CreateTargetFunc: method is nil but Provider.CreateTarget was just called")
	}
	callInfo := struct {
		S      string
		Target types.Target
	}{
		S:      s,
		Target: target,
	}
	mock.lockCreateTarget.Lock()
	mock.calls.CreateTarget = append(mock.calls.CreateTarget, callInfo)
	mock.lockCreateTarget.Unlock()
	return mock.CreateTargetFunc(s, target)
}

// CreateTargetCalls gets all the calls that were made to CreateTarget.
// Check the length with:
//     len(mockedProvider.CreateTargetCalls())
func (mock *CredsProviderMock) CreateTargetCalls() []struct {
	S      string
	Target types.Target
} {
	var calls []struct {
		S      string
		Target types.Target
	}
	mock.lockCreateTarget.RLock()
	calls = mock.calls.CreateTarget
	mock.lockCreateTarget.RUnlock()
	return calls
}

// CreateToken calls CreateTokenFunc.
func (mock *CredsProviderMock) CreateToken(s string) (types.Token, error) {
	if mock.CreateTokenFunc == nil {
		panic("CredsProviderMock.CreateTokenFunc: method is nil but Provider.CreateToken was just called")
	}
	callInfo := struct {
		S string
	}{
		S: s,
	}
	mock.lockCreateToken.Lock()
	mock.calls.CreateToken = append(mock.calls.CreateToken, callInfo)
	mock.lockCreateToken.Unlock()
	return mock.CreateTokenFunc(s)
}

// CreateTokenCalls gets all the calls that were made to CreateToken.
// Check the length with:
//     len(mockedProvider.CreateTokenCalls())
func (mock *CredsProviderMock) CreateTokenCalls() []struct {
	S string
} {
	var calls []struct {
		S string
	}
	mock.lockCreateToken.RLock()
	calls = mock.calls.CreateToken
	mock.lockCreateToken.RUnlock()
	return calls
}

// DeleteProject calls DeleteProjectFunc.
func (mock *CredsProviderMock) DeleteProject(s string) error {
	if mock.DeleteProjectFunc == nil {
		panic("CredsProviderMock.DeleteProjectFunc: method is nil but Provider.DeleteProject was just called")
	}
	callInfo := struct {
		S string
	}{
		S: s,
	}
	mock.lockDeleteProject.Lock()
	mock.calls.DeleteProject = append(mock.calls.DeleteProject, callInfo)
	mock.lockDeleteProject.Unlock()
	return mock.DeleteProjectFunc(s)
}

// DeleteProjectCalls gets all the calls that were made to DeleteProject.
// Check the length with:
//     len(mockedProvider.DeleteProjectCalls())
func (mock *CredsProviderMock) DeleteProjectCalls() []struct {
	S string
} {
	var calls []struct {
		S string
	}
	mock.lockDeleteProject.RLock()
	calls = mock.calls.DeleteProject
	mock.lockDeleteProject.RUnlock()
	return calls
}

// DeleteProjectToken calls DeleteProjectTokenFunc.
func (mock *CredsProviderMock) DeleteProjectToken(s1 string, s2 string) error {
	if mock.DeleteProjectTokenFunc == nil {
		panic("CredsProviderMock.DeleteProjectTokenFunc: method is nil but Provider.DeleteProjectToken was just called")
	}
	callInfo := struct {
		S1 string
		S2 string
	}{
		S1: s1,
		S2: s2,
	}
	mock.lockDeleteProjectToken.Lock()
	mock.calls.DeleteProjectToken = append(mock.calls.DeleteProjectToken, callInfo)
	mock.lockDeleteProjectToken.Unlock()
	return mock.DeleteProjectTokenFunc(s1, s2)
}

// DeleteProjectTokenCalls gets all the calls that were made to DeleteProjectToken.
// Check the length with:
//     len(mockedProvider.DeleteProjectTokenCalls())
func (mock *CredsProviderMock) DeleteProjectTokenCalls() []struct {
	S1 string
	S2 string
} {
	var calls []struct {
		S1 string
		S2 string
	}
	mock.lockDeleteProjectToken.RLock()
	calls = mock.calls.DeleteProjectToken
	mock.lockDeleteProjectToken.RUnlock()
	return calls
}

// DeleteTarget calls DeleteTargetFunc.
func (mock *CredsProviderMock) DeleteTarget(s1 string, s2 string) error {
	if mock.DeleteTargetFunc == nil {
		panic("CredsProviderMock.DeleteTargetFunc: method is nil but Provider.DeleteTarget was just called")
	}
	callInfo := struct {
		S1 string
		S2 string
	}{
		S1: s1,
		S2: s2,
	}
	mock.lockDeleteTarget.Lock()
	mock.calls.DeleteTarget = append(mock.calls.DeleteTarget, callInfo)
	mock.lockDeleteTarget.Unlock()
	return mock.DeleteTargetFunc(s1, s2)
}

// DeleteTargetCalls gets all the calls that were made to DeleteTarget.
// Check the length with:
//     len(mockedProvider.DeleteTargetCalls())
func (mock *CredsProviderMock) DeleteTargetCalls() []struct {
	S1 string
	S2 string
} {
	var calls []struct {
		S1 string
		S2 string
	}
	mock.lockDeleteTarget.RLock()
	calls = mock.calls.DeleteTarget
	mock.lockDeleteTarget.RUnlock()
	return calls
}

// GetProject calls GetProjectFunc.
func (mock *CredsProviderMock) GetProject(s string) (responses.GetProject, error) {
	if mock.GetProjectFunc == nil {
		panic("CredsProviderMock.GetProjectFunc: method is nil but Provider.GetProject was just called")
	}
	callInfo := struct {
		S string
	}{
		S: s,
	}
	mock.lockGetProject.Lock()
	mock.calls.GetProject = append(mock.calls.GetProject, callInfo)
	mock.lockGetProject.Unlock()
	return mock.GetProjectFunc(s)
}

// GetProjectCalls gets all the calls that were made to GetProject.
// Check the length with:
//     len(mockedProvider.GetProjectCalls())
func (mock *CredsProviderMock) GetProjectCalls() []struct {
	S string
} {
	var calls []struct {
		S string
	}
	mock.lockGetProject.RLock()
	calls = mock.calls.GetProject
	mock.lockGetProject.RUnlock()
	return calls
}

// GetProjectToken calls GetProjectTokenFunc.
func (mock *CredsProviderMock) GetProjectToken(s1 string, s2 string) (types.ProjectToken, error) {
	if mock.GetProjectTokenFunc == nil {
		panic("CredsProviderMock.GetProjectTokenFunc: method is nil but Provider.GetProjectToken was just called")
	}
	callInfo := struct {
		S1 string
		S2 string
	}{
		S1: s1,
		S2: s2,
	}
	mock.lockGetProjectToken.Lock()
	mock.calls.GetProjectToken = append(mock.calls.GetProjectToken, callInfo)
	mock.lockGetProjectToken.Unlock()
	return mock.GetProjectTokenFunc(s1, s2)
}

// GetProjectTokenCalls gets all the calls that were made to GetProjectToken.
// Check the length with:
//     len(mockedProvider.GetProjectTokenCalls())
func (mock *CredsProviderMock) GetProjectTokenCalls() []struct {
	S1 string
	S2 string
} {
	var calls []struct {
		S1 string
		S2 string
	}
	mock.lockGetProjectToken.RLock()
	calls = mock.calls.GetProjectToken
	mock.lockGetProjectToken.RUnlock()
	return calls
}

// GetTarget calls GetTargetFunc.
func (mock *CredsProviderMock) GetTarget(s1 string, s2 string) (types.Target, error) {
	if mock.GetTargetFunc == nil {
		panic("CredsProviderMock.GetTargetFunc: method is nil but Provider.GetTarget was just called")
	}
	callInfo := struct {
		S1 string
		S2 string
	}{
		S1: s1,
		S2: s2,
	}
	mock.lockGetTarget.Lock()
	mock.calls.GetTarget = append(mock.calls.GetTarget, callInfo)
	mock.lockGetTarget.Unlock()
	return mock.GetTargetFunc(s1, s2)
}

// GetTargetCalls gets all the calls that were made to GetTarget.
// Check the length with:
//     len(mockedProvider.GetTargetCalls())
func (mock *CredsProviderMock) GetTargetCalls() []struct {
	S1 string
	S2 string
} {
	var calls []struct {
		S1 string
		S2 string
	}
	mock.lockGetTarget.RLock()
	calls = mock.calls.GetTarget
	mock.lockGetTarget.RUnlock()
	return calls
}

// GetToken calls GetTokenFunc.
func (mock *CredsProviderMock) GetToken() (string, error) {
	if mock.GetTokenFunc == nil {
		panic("CredsProviderMock.GetTokenFunc: method is nil but Provider.GetToken was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetToken.Lock()
	mock.calls.GetToken = append(mock.calls.GetToken, callInfo)
	mock.lockGetToken.Unlock()
	return mock.GetTokenFunc()
}

// GetTokenCalls gets all the calls that were made to GetToken.
// Check the length with:
//     len(mockedProvider.GetTokenCalls())
func (mock *CredsProviderMock) GetTokenCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetToken.RLock()
	calls = mock.calls.GetToken
	mock.lockGetToken.RUnlock()
	return calls
}

// ListTargets calls ListTargetsFunc.
func (mock *CredsProviderMock) ListTargets(s string) ([]string, error) {
	if mock.ListTargetsFunc == nil {
		panic("CredsProviderMock.ListTargetsFunc: method is nil but Provider.ListTargets was just called")
	}
	callInfo := struct {
		S string
	}{
		S: s,
	}
	mock.lockListTargets.Lock()
	mock.calls.ListTargets = append(mock.calls.ListTargets, callInfo)
	mock.lockListTargets.Unlock()
	return mock.ListTargetsFunc(s)
}

// ListTargetsCalls gets all the calls that were made to ListTargets.
// Check the length with:
//     len(mockedProvider.ListTargetsCalls())
func (mock *CredsProviderMock) ListTargetsCalls() []struct {
	S string
} {
	var calls []struct {
		S string
	}
	mock.lockListTargets.RLock()
	calls = mock.calls.ListTargets
	mock.lockListTargets.RUnlock()
	return calls
}

// ProjectExists calls ProjectExistsFunc.
func (mock *CredsProviderMock) ProjectExists(s string) (bool, error) {
	if mock.ProjectExistsFunc == nil {
		panic("CredsProviderMock.ProjectExistsFunc: method is nil but Provider.ProjectExists was just called")
	}
	callInfo := struct {
		S string
	}{
		S: s,
	}
	mock.lockProjectExists.Lock()
	mock.calls.ProjectExists = append(mock.calls.ProjectExists, callInfo)
	mock.lockProjectExists.Unlock()
	return mock.ProjectExistsFunc(s)
}

// ProjectExistsCalls gets all the calls that were made to ProjectExists.
// Check the length with:
//     len(mockedProvider.ProjectExistsCalls())
func (mock *CredsProviderMock) ProjectExistsCalls() []struct {
	S string
} {
	var calls []struct {
		S string
	}
	mock.lockProjectExists.RLock()
	calls = mock.calls.ProjectExists
	mock.lockProjectExists.RUnlock()
	return calls
}

// TargetExists calls TargetExistsFunc.
func (mock *CredsProviderMock) TargetExists(s1 string, s2 string) (bool, error) {
	if mock.TargetExistsFunc == nil {
		panic("CredsProviderMock.TargetExistsFunc: method is nil but Provider.TargetExists was just called")
	}
	callInfo := struct {
		S1 string
		S2 string
	}{
		S1: s1,
		S2: s2,
	}
	mock.lockTargetExists.Lock()
	mock.calls.TargetExists = append(mock.calls.TargetExists, callInfo)
	mock.lockTargetExists.Unlock()
	return mock.TargetExistsFunc(s1, s2)
}

// TargetExistsCalls gets all the calls that were made to TargetExists.
// Check the length with:
//     len(mockedProvider.TargetExistsCalls())
func (mock *CredsProviderMock) TargetExistsCalls() []struct {
	S1 string
	S2 string
} {
	var calls []struct {
		S1 string
		S2 string
	}
	mock.lockTargetExists.RLock()
	calls = mock.calls.TargetExists
	mock.lockTargetExists.RUnlock()
	return calls
}

// UpdateTarget calls UpdateTargetFunc.
func (mock *CredsProviderMock) UpdateTarget(s string, target types.Target) error {
	if mock.UpdateTargetFunc == nil {
		panic("CredsProviderMock.UpdateTargetFunc: method is nil but Provider.UpdateTarget was just called")
	}
	callInfo := struct {
		S      string
		Target types.Target
	}{
		S:      s,
		Target: target,
	}
	mock.lockUpdateTarget.Lock()
	mock.calls.UpdateTarget = append(mock.calls.UpdateTarget, callInfo)
	mock.lockUpdateTarget.Unlock()
	return mock.UpdateTargetFunc(s, target)
}

// UpdateTargetCalls gets all the calls that were made to UpdateTarget.
// Check the length with:
//     len(mockedProvider.UpdateTargetCalls())
func (mock *CredsProviderMock) UpdateTargetCalls() []struct {
	S      string
	Target types.Target
} {
	var calls []struct {
		S      string
		Target types.Target
	}
	mock.lockUpdateTarget.RLock()
	calls = mock.calls.UpdateTarget
	mock.lockUpdateTarget.RUnlock()
	return calls
}
