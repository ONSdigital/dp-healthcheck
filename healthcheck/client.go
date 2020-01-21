package healthcheck

import (
	"errors"
	"sync"
)

// Client represents a health check client
type Client struct {
	Check   *Check
	Checker *Checker
	mutex   *sync.Mutex
}

// newClient returns a pointer to a new instantiated Client with
// the provided checker function and an rchttp.Clienter
func newClient(checker *Checker) (*Client, error) {
	if checker == nil {
		return nil, errors.New("expected checker but none provided")
	}

	return &Client{
		Check:   nil,
		Checker: checker,
		mutex:   &sync.Mutex{},
	}, nil
}
