package healthcheck

import (
	"context"
	"errors"
	"github.com/ONSdigital/go-ns/log"
	"time"
)

type Checker func(*context.Context) (*Check, error)

type Check struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	StatusCode  int       `json:"status_code"`
	Message     string    `json:"message"`
	LastChecked time.Time `json:"last_checked"`
	LastSuccess time.Time `json:"last_success"`
	LastFailure time.Time `json:"last_failure"`
}

type HealthCheck struct {
	Status                   string        `json:"status"`
	Version                  string        `json:"version"`
	Uptime                   time.Duration `json:"uptime"`
	StartTime                time.Time     `json:"start_time"`
	Checks                   []Check       `json:"checks"`
	Started                  bool          `json:"-"`
	Interval                 time.Duration `json:"-"`
	Clients                  []*Client     `json:"-"`
	CriticalErrorTimeout     time.Duration `json:"-"`
	TimeOfFirstCriticalError time.Time     `json:"-"`
	tickers                  []*ticker     `json:"-"`
}

// Create returns a new instantiated HealthCheck object. Caller to provide:
// version information of the app,
// criticalTimeout for how long to wait until an unhealthy dependent propagates its state to make this app unhealthy
// interval in which to check health of dependencies
// clients of type Client which contain a Checker function which is run to check the health of dependent apps
func Create(version string, criticalTimeout, interval time.Duration, clients []*Client) HealthCheck {

	hc := HealthCheck{
		Started:              false,
		Clients:              clients,
		Version:              version,
		CriticalErrorTimeout: criticalTimeout,
		Interval:             interval,
	}
	return hc
}

// AddClient adds a provided client to the healthcheck
func (hc *HealthCheck) AddClient(c *Client) {
	if hc.Started {
		log.Error(errors.New("unable to add new client, health check has already stared"), nil)
		return
	}
	hc.Clients = append(hc.Clients, c)
}

// newTickers returns an array of tickers based on the number of clients in the clients parameter.
// Each client is executed at the given interval also passed into the function
func newTickers(interval time.Duration, clients []*Client) []*ticker {
	var tickers []*ticker
	for _, client := range clients {
		tickers = append(tickers, createTicker(interval, client))
	}
	return tickers
}

// Start begins each ticker, this is used to run the health checks on dependent apps
// takes argument context and should utilise contextWithCancel
func (hc *HealthCheck) Start(ctx *context.Context) {
	hc.Started = true
	tickers := newTickers(hc.Interval, hc.Clients)
	hc.tickers = tickers
	hc.StartTime = time.Now().UTC()
	for _, ticker := range hc.tickers {
		ticker.start(*ctx)
	}
}

// Stop will cancel all tickers and thus stop all health checks
func (hc *HealthCheck) Stop() {
	for _, ticker := range hc.tickers {
		ticker.stop()
	}
}
