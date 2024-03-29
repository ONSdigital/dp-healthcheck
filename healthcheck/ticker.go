package healthcheck

import (
	"context"
	"sync"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

type ticker struct {
	timeTicker *time.Ticker
	closing    chan bool
	closed     chan bool
	check      *Check
}

// createTicker will create a ticker that calls an individual check's checker function at the provided interval
func createTicker(interval time.Duration, check *Check) *ticker {
	intervalWithJitter := calcIntervalWithJitter(interval)
	return &ticker{
		timeTicker: time.NewTicker(intervalWithJitter),
		closing:    make(chan bool),
		closed:     make(chan bool),
		check:      check,
	}
}

// start creates a goroutine to read the given ticker channel (which spins off a check for that ticker)
func (ticker *ticker) start(ctx context.Context, wg *sync.WaitGroup) {
	go func() {
		defer close(ticker.closed)

		// first run check once on application start
		wg.Add(1)
		ticker.runCheck(ctx, wg)

		for {
			select {
			case <-ctx.Done():
				ticker.stop()
			case <-ticker.closing:
				return
			case <-ticker.timeTicker.C:
				wg.Add(1)
				go ticker.runCheck(ctx, wg)
			}
		}
	}()
}

// runCheck runs a checker function of the check associated with the ticker, notifying the provided waitgroup
func (ticker *ticker) runCheck(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	err := ticker.check.checker(ctx, ticker.check.state)
	if err != nil {
		name := "no check has been made yet"
		if ticker.check.state != nil {
			name = ticker.check.state.Name()
		}
		log.Error(ctx, "failed", err, log.Data{"external_service": name})
		return
	}
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
