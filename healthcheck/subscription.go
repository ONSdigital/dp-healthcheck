package healthcheck

import (
	"sync"
)

//go:generate moq -out ./mock/subscription.go -pkg mock . Subscriber

type Subscriber interface {
	OnHealthUpdate(status string)
}

// Subscribe the provided subscriber will be notified every time that at least one of the provided checks change their state value
// The reported state will be the global state among the provided checks.
func (hc *HealthCheck) Subscribe(s Subscriber, checks ...*Check) {
	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()
	hc.subscribers[s] = checks
}

// Unsubscribe stops further notifications of health updates to the provided subscriber
func (hc *HealthCheck) Unsubscribe(s Subscriber) {
	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()
	delete(hc.subscribers, s)
}

func (hc *HealthCheck) healthChangeCallback() *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	hc.subsMutex.Lock()
	defer hc.subsMutex.Unlock()

	for s, checks := range hc.subscribers {
		status := hc.getChecksStatus(checks)
		wg.Add(1)
		go func(subscriber Subscriber) {
			defer wg.Done()
			subscriber.OnHealthUpdate(status)
		}(s)
	}

	return wg
}
