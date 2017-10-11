// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package mocks

import (
	"github.com/ONSdigital/go-ns/kafka"
	"sync"
)

var (
	lockKafkaConsumerMockIncoming sync.RWMutex
)

// KafkaConsumerMock is a mock implementation of KafkaConsumer.
//
//     func TestSomethingThatUsesKafkaConsumer(t *testing.T) {
//
//         // make and configure a mocked KafkaConsumer
//         mockedKafkaConsumer := &KafkaConsumerMock{
//             IncomingFunc: func() chan kafka.Message {
// 	               panic("TODO: mock out the Incoming method")
//             },
//         }
//
//         // TODO: use mockedKafkaConsumer in code that requires KafkaConsumer
//         //       and then make assertions.
//
//     }
type KafkaConsumerMock struct {
	// IncomingFunc mocks the Incoming method.
	IncomingFunc func() chan kafka.Message

	// calls tracks calls to the methods.
	calls struct {
		// Incoming holds details about calls to the Incoming method.
		Incoming []struct {
		}
	}
}

// Incoming calls IncomingFunc.
func (mock *KafkaConsumerMock) Incoming() chan kafka.Message {
	if mock.IncomingFunc == nil {
		panic("moq: KafkaConsumerMock.IncomingFunc is nil but KafkaConsumer.Incoming was just called")
	}
	callInfo := struct {
	}{}
	lockKafkaConsumerMockIncoming.Lock()
	mock.calls.Incoming = append(mock.calls.Incoming, callInfo)
	lockKafkaConsumerMockIncoming.Unlock()
	return mock.IncomingFunc()
}

// IncomingCalls gets all the calls that were made to Incoming.
// Check the length with:
//     len(mockedKafkaConsumer.IncomingCalls())
func (mock *KafkaConsumerMock) IncomingCalls() []struct {
} {
	var calls []struct {
	}
	lockKafkaConsumerMockIncoming.RLock()
	calls = mock.calls.Incoming
	lockKafkaConsumerMockIncoming.RUnlock()
	return calls
}

var (
	lockKafkaMessageMockCommit  sync.RWMutex
	lockKafkaMessageMockGetData sync.RWMutex
)

// KafkaMessageMock is a mock implementation of KafkaMessage.
//
//     func TestSomethingThatUsesKafkaMessage(t *testing.T) {
//
//         // make and configure a mocked KafkaMessage
//         mockedKafkaMessage := &KafkaMessageMock{
//             CommitFunc: func()  {
// 	               panic("TODO: mock out the Commit method")
//             },
//             GetDataFunc: func() []byte {
// 	               panic("TODO: mock out the GetData method")
//             },
//         }
//
//         // TODO: use mockedKafkaMessage in code that requires KafkaMessage
//         //       and then make assertions.
//
//     }
type KafkaMessageMock struct {
	// CommitFunc mocks the Commit method.
	CommitFunc func()

	// GetDataFunc mocks the GetData method.
	GetDataFunc func() []byte

	// calls tracks calls to the methods.
	calls struct {
		// Commit holds details about calls to the Commit method.
		Commit []struct {
		}
		// GetData holds details about calls to the GetData method.
		GetData []struct {
		}
	}
}

// Commit calls CommitFunc.
func (mock *KafkaMessageMock) Commit() {
	if mock.CommitFunc == nil {
		panic("moq: KafkaMessageMock.CommitFunc is nil but KafkaMessage.Commit was just called")
	}
	callInfo := struct {
	}{}
	lockKafkaMessageMockCommit.Lock()
	mock.calls.Commit = append(mock.calls.Commit, callInfo)
	lockKafkaMessageMockCommit.Unlock()
	mock.CommitFunc()
}

// CommitCalls gets all the calls that were made to Commit.
// Check the length with:
//     len(mockedKafkaMessage.CommitCalls())
func (mock *KafkaMessageMock) CommitCalls() []struct {
} {
	var calls []struct {
	}
	lockKafkaMessageMockCommit.RLock()
	calls = mock.calls.Commit
	lockKafkaMessageMockCommit.RUnlock()
	return calls
}

// GetData calls GetDataFunc.
func (mock *KafkaMessageMock) GetData() []byte {
	if mock.GetDataFunc == nil {
		panic("moq: KafkaMessageMock.GetDataFunc is nil but KafkaMessage.GetData was just called")
	}
	callInfo := struct {
	}{}
	lockKafkaMessageMockGetData.Lock()
	mock.calls.GetData = append(mock.calls.GetData, callInfo)
	lockKafkaMessageMockGetData.Unlock()
	return mock.GetDataFunc()
}

// GetDataCalls gets all the calls that were made to GetData.
// Check the length with:
//     len(mockedKafkaMessage.GetDataCalls())
func (mock *KafkaMessageMock) GetDataCalls() []struct {
} {
	var calls []struct {
	}
	lockKafkaMessageMockGetData.RLock()
	calls = mock.calls.GetData
	lockKafkaMessageMockGetData.RUnlock()
	return calls
}

var (
	lockReceiverMockProcessMessage sync.RWMutex
)

// ReceiverMock is a mock implementation of Receiver.
//
//     func TestSomethingThatUsesReceiver(t *testing.T) {
//
//         // make and configure a mocked Receiver
//         mockedReceiver := &ReceiverMock{
//             ProcessMessageFunc: func(event kafka.Message) error {
// 	               panic("TODO: mock out the ProcessMessage method")
//             },
//         }
//
//         // TODO: use mockedReceiver in code that requires Receiver
//         //       and then make assertions.
//
//     }
type ReceiverMock struct {
	// ProcessMessageFunc mocks the ProcessMessage method.
	ProcessMessageFunc func(event kafka.Message) error

	// calls tracks calls to the methods.
	calls struct {
		// ProcessMessage holds details about calls to the ProcessMessage method.
		ProcessMessage []struct {
			// Event is the event argument value.
			Event kafka.Message
		}
	}
}

// ProcessMessage calls ProcessMessageFunc.
func (mock *ReceiverMock) ProcessMessage(event kafka.Message) error {
	if mock.ProcessMessageFunc == nil {
		panic("moq: ReceiverMock.ProcessMessageFunc is nil but Receiver.ProcessMessage was just called")
	}
	callInfo := struct {
		Event kafka.Message
	}{
		Event: event,
	}
	lockReceiverMockProcessMessage.Lock()
	mock.calls.ProcessMessage = append(mock.calls.ProcessMessage, callInfo)
	lockReceiverMockProcessMessage.Unlock()
	return mock.ProcessMessageFunc(event)
}

// ProcessMessageCalls gets all the calls that were made to ProcessMessage.
// Check the length with:
//     len(mockedReceiver.ProcessMessageCalls())
func (mock *ReceiverMock) ProcessMessageCalls() []struct {
	Event kafka.Message
} {
	var calls []struct {
		Event kafka.Message
	}
	lockReceiverMockProcessMessage.RLock()
	calls = mock.calls.ProcessMessage
	lockReceiverMockProcessMessage.RUnlock()
	return calls
}