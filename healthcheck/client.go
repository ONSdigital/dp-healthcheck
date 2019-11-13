package healthcheck

import (
	rchttp "github.com/ONSdigital/dp-rchttp"
	"sync"
)

type Client struct {
	Clienter   rchttp.Clienter
	Check      *Check
	Checker    *Checker
	MutexCheck *sync.Mutex
}

// NewClient returns a pointer to a new instantiated Client with
// the provided checker function and an optional rchttp.Clienter
func NewClient(cli rchttp.Clienter, checker *Checker) *Client {
	if cli == nil {
		cli = rchttp.NewClient()
	}

	return &Client{
		Clienter:   cli,
		Check:      nil,
		Checker:    checker,
		MutexCheck: &sync.Mutex{},
	}
}
