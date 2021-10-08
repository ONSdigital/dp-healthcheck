// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"sync"
)

var (
	lockSubscriberMockOnHealthUpdate sync.RWMutex
)

// SubscriberMock is a mock implementation of healthcheck.Subscriber.
//
//     func TestSomethingThatUsesSubscriber(t *testing.T) {
//
//         // make and configure a mocked healthcheck.Subscriber
//         mockedSubscriber := &SubscriberMock{
//             OnHealthUpdateFunc: func(status string)  {
// 	               panic("mock out the OnHealthUpdate method")
//             },
//         }
//
//         // use mockedSubscriber in code that requires healthcheck.Subscriber
//         // and then make assertions.
//
//     }
type SubscriberMock struct {
	// OnHealthUpdateFunc mocks the OnHealthUpdate method.
	OnHealthUpdateFunc func(status string)

	// calls tracks calls to the methods.
	calls struct {
		// OnHealthUpdate holds details about calls to the OnHealthUpdate method.
		OnHealthUpdate []struct {
			// Status is the status argument value.
			Status string
		}
	}
}

// OnHealthUpdate calls OnHealthUpdateFunc.
func (mock *SubscriberMock) OnHealthUpdate(status string) {
	if mock.OnHealthUpdateFunc == nil {
		panic("SubscriberMock.OnHealthUpdateFunc: method is nil but Subscriber.OnHealthUpdate was just called")
	}
	callInfo := struct {
		Status string
	}{
		Status: status,
	}
	lockSubscriberMockOnHealthUpdate.Lock()
	mock.calls.OnHealthUpdate = append(mock.calls.OnHealthUpdate, callInfo)
	lockSubscriberMockOnHealthUpdate.Unlock()
	mock.OnHealthUpdateFunc(status)
}

// OnHealthUpdateCalls gets all the calls that were made to OnHealthUpdate.
// Check the length with:
//     len(mockedSubscriber.OnHealthUpdateCalls())
func (mock *SubscriberMock) OnHealthUpdateCalls() []struct {
	Status string
} {
	var calls []struct {
		Status string
	}
	lockSubscriberMockOnHealthUpdate.RLock()
	calls = mock.calls.OnHealthUpdate
	lockSubscriberMockOnHealthUpdate.RUnlock()
	return calls
}
