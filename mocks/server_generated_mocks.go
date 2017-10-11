// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package mocks

import (
	"sync"
)

var (
	lockClearableCacheMockClear sync.RWMutex
)

// ClearableCacheMock is a mock implementation of ClearableCache.
//
//     func TestSomethingThatUsesClearableCache(t *testing.T) {
//
//         // make and configure a mocked ClearableCache
//         mockedClearableCache := &ClearableCacheMock{
//             ClearFunc: func()  {
// 	               panic("TODO: mock out the Clear method")
//             },
//         }
//
//         // TODO: use mockedClearableCache in code that requires ClearableCache
//         //       and then make assertions.
//
//     }
type ClearableCacheMock struct {
	// ClearFunc mocks the Clear method.
	ClearFunc func()

	// calls tracks calls to the methods.
	calls struct {
		// Clear holds details about calls to the Clear method.
		Clear []struct {
		}
	}
}

// Clear calls ClearFunc.
func (mock *ClearableCacheMock) Clear() {
	if mock.ClearFunc == nil {
		panic("moq: ClearableCacheMock.ClearFunc is nil but ClearableCache.Clear was just called")
	}
	callInfo := struct {
	}{}
	lockClearableCacheMockClear.Lock()
	mock.calls.Clear = append(mock.calls.Clear, callInfo)
	lockClearableCacheMockClear.Unlock()
	mock.ClearFunc()
}

// ClearCalls gets all the calls that were made to Clear.
// Check the length with:
//     len(mockedClearableCache.ClearCalls())
func (mock *ClearableCacheMock) ClearCalls() []struct {
} {
	var calls []struct {
	}
	lockClearableCacheMockClear.RLock()
	calls = mock.calls.Clear
	lockClearableCacheMockClear.RUnlock()
	return calls
}