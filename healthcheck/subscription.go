package healthcheck

import (
	"sync"
)

//go:generate moq -out ./mock/subscription.go -pkg mock . Subscriber

type Subscriber interface {
	OnHealthUpdate(status string)
}

// Subscribe will subscribe the subscriber to the provided checks.
// This method may be called multiple times to subscribe to more checks and it is idempotent.
// The subscriber will be notified of the accumulated state of the subscribed Checks every time that a check changes its state.
func (hc *HealthCheck) Subscribe(s Subscriber, checks ...*Check) {
	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()

	// if the subscriber was not subscribed yet, create an entry for it
	_, ok := hc.subscribers[s]
	if !ok {
		hc.subscribers[s] = map[*Check]struct{}{}
	}

	// add all checkers (using a struct{} map for uniqueness and efficiency)
	for _, check := range checks {
		hc.subscribers[s][check] = struct{}{}
	}
}

// SubscribeAll will subscribe the subscriber to all the Checks that have been added.
// The subscriber will be notified of the global state every time that a check changes its state.
func (hc *HealthCheck) SubscribeAll(s Subscriber) {
	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()

	hc.subscribers[s] = map[*Check]struct{}{}
	for _, check := range hc.Checks {
		hc.subscribers[s][check] = struct{}{}
	}
}

// Unsubscribe removes the provided checks that will be used in order to determine the accumulated state for the provided subscriber
func (hc *HealthCheck) Unsubscribe(s Subscriber, checks ...*Check) {
	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()

	// if the subscriber was not subscribed, there is nothing to do
	subscribed, ok := hc.subscribers[s]
	if !ok {
		return
	}

	// remove all checkers, if present
	for _, check := range checks {
		delete(subscribed, check)
	}

	// if the subscriber is empty, it is removed.
	if len(subscribed) == 0 {
		delete(hc.subscribers, s)
	}
}

// UnsubscribeAll stops further notifications of health updates to the provided subscriber
func (hc *HealthCheck) UnsubscribeAll(s Subscriber) {
	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()
	delete(hc.subscribers, s)
}

func (hc *HealthCheck) healthChangeCallback() *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()

	for s, checks := range hc.subscribers {
		checkList := []*Check{}
		for check := range checks {
			checkList = append(checkList, check)
		}
		status := hc.getChecksStatus(checkList)
		wg.Add(1)
		go func(subscriber Subscriber) {
			defer wg.Done()
			subscriber.OnHealthUpdate(status)
		}(s)
	}

	return wg
}
