package healthcheck

import (
	"errors"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"sync"
)

type Client struct {
	Clienter rchttp.Clienter
	Check    *Check
	Checker  *Checker
	mutex    *sync.Mutex
}

// NewClient returns a pointer to a new instantiated Client with
// the provided checker function and an optional rchttp.Clienter
func NewClient(cli rchttp.Clienter, checker *Checker) (*Client, error) {
	if cli == nil {
		return nil, errors.New("expected clienter but none provided")
	}

	return &Client{
		Clienter: cli,
		Check:    nil,
		Checker:  checker,
		mutex:    &sync.Mutex{},
	}, nil
}
