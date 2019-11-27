package healthcheck

import (
	"context"
	"github.com/ONSdigital/go-ns/log"
	"time"
)

type ticker struct {
	timeTicker *time.Ticker
	closing    chan bool
	closed     chan bool
	client     *Client
}

// createTicker will create a ticker that calls an individual clients checker function at the provided interval
func createTicker(interval time.Duration, client *Client) *ticker {
	intervalWithJitter := calcIntervalWithJitter(interval)
	ticker := ticker{
		timeTicker: time.NewTicker(intervalWithJitter),
		closing:    make(chan bool),
		closed:     make(chan bool),
		client:     client,
	}
	return &ticker
}

// start creates a goroutine to read the given ticker channel (which spins off a check for that ticker)
func (ticker ticker) start(ctx context.Context) {
	go func() {
		defer close(ticker.closed)
		for {
			select {
			case <-ctx.Done():
				ticker.stop()
			case <-ticker.closing:
				return
			case <-ticker.timeTicker.C:
				go ticker.runCheck(ctx)
			}
		}
	}()
}

// runCheck runs a checker function on a single client associated to the ticker
func (ticker ticker) runCheck(ctx context.Context) {
	checker := *ticker.client.Checker
	checkResults, err := checker(&ctx)
	if err != nil {
		// If first check has failed then there is no way to know which app it was attempting to check
		if ticker.client.Check != nil {
			log.Error(err, log.Data{"external_service": ticker.client.Check.Name})
		} else {
			log.Error(err, nil)
		}
		return
	}

	ticker.client.mutex.Lock()
	defer ticker.client.mutex.Unlock()
	ticker.client.Check = checkResults
}

// stop the ticker
func (ticker ticker) stop() {
	ticker.timeTicker.Stop()
	close(ticker.closing)
	<-ticker.closed
}
