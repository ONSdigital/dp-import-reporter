// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package mocks

import (
	"github.com/ONSdigital/dp-import-reporter/model"
	"sync"
)

var (
	lockDatasetAPICliMockAddEventToInstance   sync.RWMutex
	lockDatasetAPICliMockGetInstance          sync.RWMutex
	lockDatasetAPICliMockUpdateInstanceStatus sync.RWMutex
)

// DatasetAPICliMock is a mock implementation of DatasetAPICli.
//
//     func TestSomethingThatUsesDatasetAPICli(t *testing.T) {
//
//         // make and configure a mocked DatasetAPICli
//         mockedDatasetAPICli := &DatasetAPICliMock{
//             AddEventToInstanceFunc: func(instanceID string, e *model.Event) error {
// 	               panic("TODO: mock out the AddEventToInstance method")
//             },
//             GetInstanceFunc: func(instanceID string) (*model.Instance, error) {
// 	               panic("TODO: mock out the GetInstance method")
//             },
//             UpdateInstanceStatusFunc: func(instanceID string, state *model.State) error {
// 	               panic("TODO: mock out the UpdateInstanceStatus method")
//             },
//         }
//
//         // TODO: use mockedDatasetAPICli in code that requires DatasetAPICli
//         //       and then make assertions.
//
//     }
type DatasetAPICliMock struct {
	// AddEventToInstanceFunc mocks the AddEventToInstance method.
	AddEventToInstanceFunc func(instanceID string, e *model.Event) error

	// GetInstanceFunc mocks the GetInstance method.
	GetInstanceFunc func(instanceID string) (*model.Instance, error)

	// UpdateInstanceStatusFunc mocks the UpdateInstanceStatus method.
	UpdateInstanceStatusFunc func(instanceID string, state *model.State) error

	// calls tracks calls to the methods.
	calls struct {
		// AddEventToInstance holds details about calls to the AddEventToInstance method.
		AddEventToInstance []struct {
			// InstanceID is the instanceID argument value.
			InstanceID string
			// E is the e argument value.
			E *model.Event
		}
		// GetInstance holds details about calls to the GetInstance method.
		GetInstance []struct {
			// InstanceID is the instanceID argument value.
			InstanceID string
		}
		// UpdateInstanceStatus holds details about calls to the UpdateInstanceStatus method.
		UpdateInstanceStatus []struct {
			// InstanceID is the instanceID argument value.
			InstanceID string
			// State is the state argument value.
			State *model.State
		}
	}
}

// AddEventToInstance calls AddEventToInstanceFunc.
func (mock *DatasetAPICliMock) AddEventToInstance(instanceID string, e *model.Event) error {
	if mock.AddEventToInstanceFunc == nil {
		panic("moq: DatasetAPICliMock.AddEventToInstanceFunc is nil but DatasetAPICli.AddEventToInstance was just called")
	}
	callInfo := struct {
		InstanceID string
		E          *model.Event
	}{
		InstanceID: instanceID,
		E:          e,
	}
	lockDatasetAPICliMockAddEventToInstance.Lock()
	mock.calls.AddEventToInstance = append(mock.calls.AddEventToInstance, callInfo)
	lockDatasetAPICliMockAddEventToInstance.Unlock()
	return mock.AddEventToInstanceFunc(instanceID, e)
}

// AddEventToInstanceCalls gets all the calls that were made to AddEventToInstance.
// Check the length with:
//     len(mockedDatasetAPICli.AddEventToInstanceCalls())
func (mock *DatasetAPICliMock) AddEventToInstanceCalls() []struct {
	InstanceID string
	E          *model.Event
} {
	var calls []struct {
		InstanceID string
		E          *model.Event
	}
	lockDatasetAPICliMockAddEventToInstance.RLock()
	calls = mock.calls.AddEventToInstance
	lockDatasetAPICliMockAddEventToInstance.RUnlock()
	return calls
}

// GetInstance calls GetInstanceFunc.
func (mock *DatasetAPICliMock) GetInstance(instanceID string) (*model.Instance, error) {
	if mock.GetInstanceFunc == nil {
		panic("moq: DatasetAPICliMock.GetInstanceFunc is nil but DatasetAPICli.GetInstance was just called")
	}
	callInfo := struct {
		InstanceID string
	}{
		InstanceID: instanceID,
	}
	lockDatasetAPICliMockGetInstance.Lock()
	mock.calls.GetInstance = append(mock.calls.GetInstance, callInfo)
	lockDatasetAPICliMockGetInstance.Unlock()
	return mock.GetInstanceFunc(instanceID)
}

// GetInstanceCalls gets all the calls that were made to GetInstance.
// Check the length with:
//     len(mockedDatasetAPICli.GetInstanceCalls())
func (mock *DatasetAPICliMock) GetInstanceCalls() []struct {
	InstanceID string
} {
	var calls []struct {
		InstanceID string
	}
	lockDatasetAPICliMockGetInstance.RLock()
	calls = mock.calls.GetInstance
	lockDatasetAPICliMockGetInstance.RUnlock()
	return calls
}

// UpdateInstanceStatus calls UpdateInstanceStatusFunc.
func (mock *DatasetAPICliMock) UpdateInstanceStatus(instanceID string, state *model.State) error {
	if mock.UpdateInstanceStatusFunc == nil {
		panic("moq: DatasetAPICliMock.UpdateInstanceStatusFunc is nil but DatasetAPICli.UpdateInstanceStatus was just called")
	}
	callInfo := struct {
		InstanceID string
		State      *model.State
	}{
		InstanceID: instanceID,
		State:      state,
	}
	lockDatasetAPICliMockUpdateInstanceStatus.Lock()
	mock.calls.UpdateInstanceStatus = append(mock.calls.UpdateInstanceStatus, callInfo)
	lockDatasetAPICliMockUpdateInstanceStatus.Unlock()
	return mock.UpdateInstanceStatusFunc(instanceID, state)
}

// UpdateInstanceStatusCalls gets all the calls that were made to UpdateInstanceStatus.
// Check the length with:
//     len(mockedDatasetAPICli.UpdateInstanceStatusCalls())
func (mock *DatasetAPICliMock) UpdateInstanceStatusCalls() []struct {
	InstanceID string
	State      *model.State
} {
	var calls []struct {
		InstanceID string
		State      *model.State
	}
	lockDatasetAPICliMockUpdateInstanceStatus.RLock()
	calls = mock.calls.UpdateInstanceStatus
	lockDatasetAPICliMockUpdateInstanceStatus.RUnlock()
	return calls
}

var (
	lockCacheMockDel sync.RWMutex
	lockCacheMockGet sync.RWMutex
	lockCacheMockSet sync.RWMutex
	lockCacheMockTTL sync.RWMutex
)

// CacheMock is a mock implementation of Cache.
//
//     func TestSomethingThatUsesCache(t *testing.T) {
//
//         // make and configure a mocked Cache
//         mockedCache := &CacheMock{
//             DelFunc: func(key []byte) bool {
// 	               panic("TODO: mock out the Del method")
//             },
//             GetFunc: func(key []byte) ([]byte, error) {
// 	               panic("TODO: mock out the Get method")
//             },
//             SetFunc: func(key []byte, value []byte, expireSeconds int) error {
// 	               panic("TODO: mock out the Set method")
//             },
//             TTLFunc: func(key []byte) (uint32, error) {
// 	               panic("TODO: mock out the TTL method")
//             },
//         }
//
//         // TODO: use mockedCache in code that requires Cache
//         //       and then make assertions.
//
//     }
type CacheMock struct {
	// DelFunc mocks the Del method.
	DelFunc func(key []byte) bool

	// GetFunc mocks the Get method.
	GetFunc func(key []byte) ([]byte, error)

	// SetFunc mocks the Set method.
	SetFunc func(key []byte, value []byte, expireSeconds int) error

	// TTLFunc mocks the TTL method.
	TTLFunc func(key []byte) (uint32, error)

	// calls tracks calls to the methods.
	calls struct {
		// Del holds details about calls to the Del method.
		Del []struct {
			// Key is the key argument value.
			Key []byte
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// Key is the key argument value.
			Key []byte
		}
		// Set holds details about calls to the Set method.
		Set []struct {
			// Key is the key argument value.
			Key []byte
			// Value is the value argument value.
			Value []byte
			// ExpireSeconds is the expireSeconds argument value.
			ExpireSeconds int
		}
		// TTL holds details about calls to the TTL method.
		TTL []struct {
			// Key is the key argument value.
			Key []byte
		}
	}
}

// Del calls DelFunc.
func (mock *CacheMock) Del(key []byte) bool {
	if mock.DelFunc == nil {
		panic("moq: CacheMock.DelFunc is nil but Cache.Del was just called")
	}
	callInfo := struct {
		Key []byte
	}{
		Key: key,
	}
	lockCacheMockDel.Lock()
	mock.calls.Del = append(mock.calls.Del, callInfo)
	lockCacheMockDel.Unlock()
	return mock.DelFunc(key)
}

// DelCalls gets all the calls that were made to Del.
// Check the length with:
//     len(mockedCache.DelCalls())
func (mock *CacheMock) DelCalls() []struct {
	Key []byte
} {
	var calls []struct {
		Key []byte
	}
	lockCacheMockDel.RLock()
	calls = mock.calls.Del
	lockCacheMockDel.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *CacheMock) Get(key []byte) ([]byte, error) {
	if mock.GetFunc == nil {
		panic("moq: CacheMock.GetFunc is nil but Cache.Get was just called")
	}
	callInfo := struct {
		Key []byte
	}{
		Key: key,
	}
	lockCacheMockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	lockCacheMockGet.Unlock()
	return mock.GetFunc(key)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//     len(mockedCache.GetCalls())
func (mock *CacheMock) GetCalls() []struct {
	Key []byte
} {
	var calls []struct {
		Key []byte
	}
	lockCacheMockGet.RLock()
	calls = mock.calls.Get
	lockCacheMockGet.RUnlock()
	return calls
}

// Set calls SetFunc.
func (mock *CacheMock) Set(key []byte, value []byte, expireSeconds int) error {
	if mock.SetFunc == nil {
		panic("moq: CacheMock.SetFunc is nil but Cache.Set was just called")
	}
	callInfo := struct {
		Key           []byte
		Value         []byte
		ExpireSeconds int
	}{
		Key:           key,
		Value:         value,
		ExpireSeconds: expireSeconds,
	}
	lockCacheMockSet.Lock()
	mock.calls.Set = append(mock.calls.Set, callInfo)
	lockCacheMockSet.Unlock()
	return mock.SetFunc(key, value, expireSeconds)
}

// SetCalls gets all the calls that were made to Set.
// Check the length with:
//     len(mockedCache.SetCalls())
func (mock *CacheMock) SetCalls() []struct {
	Key           []byte
	Value         []byte
	ExpireSeconds int
} {
	var calls []struct {
		Key           []byte
		Value         []byte
		ExpireSeconds int
	}
	lockCacheMockSet.RLock()
	calls = mock.calls.Set
	lockCacheMockSet.RUnlock()
	return calls
}

// TTL calls TTLFunc.
func (mock *CacheMock) TTL(key []byte) (uint32, error) {
	if mock.TTLFunc == nil {
		panic("moq: CacheMock.TTLFunc is nil but Cache.TTL was just called")
	}
	callInfo := struct {
		Key []byte
	}{
		Key: key,
	}
	lockCacheMockTTL.Lock()
	mock.calls.TTL = append(mock.calls.TTL, callInfo)
	lockCacheMockTTL.Unlock()
	return mock.TTLFunc(key)
}

// TTLCalls gets all the calls that were made to TTL.
// Check the length with:
//     len(mockedCache.TTLCalls())
func (mock *CacheMock) TTLCalls() []struct {
	Key []byte
} {
	var calls []struct {
		Key []byte
	}
	lockCacheMockTTL.RLock()
	calls = mock.calls.TTL
	lockCacheMockTTL.RUnlock()
	return calls
}

var (
	lockEventHandlerMockHandleEvent sync.RWMutex
)

// EventHandlerMock is a mock implementation of EventHandler.
//
//     func TestSomethingThatUsesEventHandler(t *testing.T) {
//
//         // make and configure a mocked EventHandler
//         mockedEventHandler := &EventHandlerMock{
//             HandleEventFunc: func(e *model.ReportEvent) error {
// 	               panic("TODO: mock out the HandleEvent method")
//             },
//         }
//
//         // TODO: use mockedEventHandler in code that requires EventHandler
//         //       and then make assertions.
//
//     }
type EventHandlerMock struct {
	// HandleEventFunc mocks the HandleEvent method.
	HandleEventFunc func(e *model.ReportEvent) error

	// calls tracks calls to the methods.
	calls struct {
		// HandleEvent holds details about calls to the HandleEvent method.
		HandleEvent []struct {
			// E is the e argument value.
			E *model.ReportEvent
		}
	}
}

// HandleEvent calls HandleEventFunc.
func (mock *EventHandlerMock) HandleEvent(e *model.ReportEvent) error {
	if mock.HandleEventFunc == nil {
		panic("moq: EventHandlerMock.HandleEventFunc is nil but EventHandler.HandleEvent was just called")
	}
	callInfo := struct {
		E *model.ReportEvent
	}{
		E: e,
	}
	lockEventHandlerMockHandleEvent.Lock()
	mock.calls.HandleEvent = append(mock.calls.HandleEvent, callInfo)
	lockEventHandlerMockHandleEvent.Unlock()
	return mock.HandleEventFunc(e)
}

// HandleEventCalls gets all the calls that were made to HandleEvent.
// Check the length with:
//     len(mockedEventHandler.HandleEventCalls())
func (mock *EventHandlerMock) HandleEventCalls() []struct {
	E *model.ReportEvent
} {
	var calls []struct {
		E *model.ReportEvent
	}
	lockEventHandlerMockHandleEvent.RLock()
	calls = mock.calls.HandleEvent
	lockEventHandlerMockHandleEvent.RUnlock()
	return calls
}
