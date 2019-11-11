package HealthCheck

import (
	"context"
	"github.com/ONSdigital/go-ns/log"
	"time"
)
type ticker struct {
	TimeTicker *time.Ticker
	Closing    chan bool
	Closed     chan bool
	Client     *Client
}

// createTicker will create a ticker that calls an individual clients checker function at the provided interval
func createTicker(interval time.Duration, client *Client) *ticker {
	intervalWithJitter := calcIntervalWithJitter(interval)
	ticker := ticker{
		TimeTicker: time.NewTicker(intervalWithJitter),
		Closing:    make(chan bool),
		Closed:     make(chan bool),
		Client:     client,
	}
	return &ticker
}

// start opens channels for ticker processes to run on, also used to close requests
func (ticker ticker) start(ctx context.Context) {
	go func() {
		defer close(ticker.Closed)
		for {
			select {
			case <- ctx.Done():
				ticker.stop()
			case <-ticker.Closing:
				return
			case <-ticker.TimeTicker.C:
				go ticker.runCheck(ctx)
			}
		}
	}()
}

// runCheck runs a checker function on a single client associated to the ticker
func (ticker ticker) runCheck(ctx context.Context) {
	checker := *ticker.Client.Checker
	checkResults, err := checker.CheckAppHealth(&ctx)
	if err != nil {
		log.ErrorC("unsuccessful Health check", err, log.Data{"external_service": ticker.Client.Check.Name})
	} else {
		ticker.Client.MutexCheck.Lock()
		ticker.Client.Check = checkResults
		ticker.Client.MutexCheck.Unlock()
	}
}

// stop the ticker
func (ticker ticker) stop() {
	ticker.TimeTicker.Stop()
	close(ticker.Closing)
	<-ticker.Closed
}
