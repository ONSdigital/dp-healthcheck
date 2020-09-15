package healthcheck

import (
	"context"
	"sync"
	"time"

	"github.com/ONSdigital/log.go/log"
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

		const maxChecks int = 5 // Max number of healthchecks in flight
		var checkInFlight int
		checkDone := make(chan bool, maxChecks)

		for {
			//fmt.Printf("count: %v\n", checkInFlight)
			select {
			case <-ctx.Done():
				ticker.stop()
			case <-ticker.closing:
				close(checkDone)
				return
			case <-ticker.timeTicker.C:
				if checkInFlight < maxChecks {
					checkInFlight++
					wg.Add(1)
					go ticker.runCheck(ctx, wg, checkDone)
				}
			case <-checkDone:
				checkInFlight--
			}
		}
	}()
}

// runCheck runs a checker function of the check associated with the ticker, notifying the provided waitgroup
func (ticker *ticker) runCheck(ctx context.Context, wg *sync.WaitGroup, done chan bool) {

	defer func() {
		if x := recover(); x != nil {
			// do nothing ... just handle timing corner case and avoid "panic: send on closed channel"
			//fmt.Printf("in recover of runCheck : %v\n", x)
		}
	}()

	defer func() { // this gets called before the defer above
		wg.Done()
		done <- true
	}()

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

// stop the ticker
func (ticker *ticker) stop() {
	if ticker.isStopping() {
		return
	}
	ticker.timeTicker.Stop()

	defer func() {
		if x := recover(); x != nil {
			// do nothing ... just handle timing corner case and avoid "panic: send on closed channel"
			//fmt.Printf("in recover of stop : %v\n", x)
		}
		<-ticker.closed
		//fmt.Printf("got ticker.closed\n")
	}()

	close(ticker.closing)
}

func (ticker *ticker) isStopping() bool {
	select {
	case <-ticker.closing:
		return true
	default:
	}
	return false
}
