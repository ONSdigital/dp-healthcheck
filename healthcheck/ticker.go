package healthcheck

import (
	"context"
	"time"

	"github.com/ONSdigital/log.go/log"
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
	return &ticker{
		timeTicker: time.NewTicker(intervalWithJitter),
		closing:    make(chan bool),
		closed:     make(chan bool),
		client:     client,
	}
}

// start creates a goroutine to read the given ticker channel (which spins off a check for that ticker)
func (ticker *ticker) start(ctx context.Context) {
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

// runCheck runs a checker function on the client associated with the ticker
func (ticker *ticker) runCheck(ctx context.Context) {
	checkerFunc := *ticker.client.Checker
	checkResults, err := checkerFunc(ctx)
	if err != nil {
		name := "no check has been made yet"
		if ticker.client.Check != nil {
			name = ticker.client.Check.Name
		}
		log.Event(nil, "failed", log.Error(err), log.Data{"external_service": name})
		return
	}

	ticker.client.mutex.Lock()
	defer ticker.client.mutex.Unlock()
	ticker.client.Check = checkResults
}

// stop the ticker
func (ticker *ticker) stop() {
	if ticker.isStopping() {
		return
	}
	ticker.timeTicker.Stop()
	close(ticker.closing)
	<-ticker.closed
}

func (ticker *ticker) isStopping() bool {
	select {
	case <-ticker.closing:
		return true
	default:
	}
	return false
}
