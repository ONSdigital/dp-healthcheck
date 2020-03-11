package healthcheck

import (
	"context"
	"time"

	gonslog "github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/log.go/log"
)

type ticker struct {
	timeTicker *time.Ticker
	check      *Check
}

// createTicker will create a ticker that calls an individual check's checker function at the provided interval
func createTicker(interval time.Duration, check *Check) *ticker {
	intervalWithJitter := calcIntervalWithJitter(interval)
	return &ticker{
		timeTicker: time.NewTicker(intervalWithJitter),
		check:      check,
	}
}

// start creates a goroutine to read the given ticker channel (which spins off a check for that ticker)
func (ticker *ticker) start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.timeTicker.Stop()
				gonslog.Debug("ticker.go : ctx.Done() and ticker stopped", nil) // TODO: remove when / if happy code runs OK
				return
			case <-ticker.timeTicker.C:
				go ticker.runCheck(ctx)
			}
		}
	}()
}

// runCheck runs a checker function of the check associated with the ticker
func (ticker *ticker) runCheck(ctx context.Context) {
	err := ticker.check.checker(ctx, ticker.check.state)
	if err != nil {
		name := "no check has been made yet"
		if ticker.check.state != nil {
			name = ticker.check.state.Name()
		}
		log.Event(nil, "failed", log.Error(err), log.Data{"external_service": name})
		return
	}
}
