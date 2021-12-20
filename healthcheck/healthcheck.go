package healthcheck

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

const language = "go"

// HealthCheck represents the app's health check, including its component checks
type HealthCheck struct {
	Status                   string        `json:"status"`
	Version                  VersionInfo   `json:"version"`
	Uptime                   time.Duration `json:"uptime"`
	StartTime                time.Time     `json:"start_time"`
	Checks                   []*Check      `json:"checks"`
	interval                 time.Duration
	criticalErrorTimeout     time.Duration
	timeOfFirstCriticalError time.Time
	tickers                  []*ticker
	context                  context.Context
	tickersWaitgroup         *sync.WaitGroup
	statusLock               *sync.RWMutex
	subscribers              map[Subscriber]map[*Check]struct{}
	subsMutex                *sync.Mutex
	stopper                  chan struct{}
}

// VersionInfo represents the version information of an app
type VersionInfo struct {
	BuildTime       time.Time `json:"build_time"`
	GitCommit       string    `json:"git_commit"`
	Language        string    `json:"language"`
	LanguageVersion string    `json:"language_version"`
	Version         string    `json:"version"`
}

// New returns a new instantiated HealthCheck object. Caller to provide:
// version information of the app,
// criticalTimeout for how long to wait until an unhealthy dependent propagates its state to make this app unhealthy
// interval in which to check health of dependencies
func New(version VersionInfo, criticalTimeout, interval time.Duration) HealthCheck {
	return HealthCheck{
		Checks:               []*Check{},
		Version:              version,
		criticalErrorTimeout: criticalTimeout,
		interval:             interval,
		tickers:              []*ticker{},
		tickersWaitgroup:     &sync.WaitGroup{},
		statusLock:           &sync.RWMutex{},
		subscribers:          map[Subscriber]map[*Check]struct{}{},
		subsMutex:            &sync.Mutex{},
	}
}

// NewVersionInfo returns a health check version info object. Caller to provide:
// buildTime for when the app was built as a unix time stamp in string form
// gitCommit the SHA-1 commit hash of the built app
// version the semantic version of the built app
func NewVersionInfo(buildTime, gitCommit, version string) (VersionInfo, error) {
	versionInfo := VersionInfo{
		BuildTime:       time.Unix(0, 0),
		GitCommit:       gitCommit,
		Language:        language,
		LanguageVersion: runtime.Version(),
		Version:         version,
	}

	parsedBuildTime, err := strconv.ParseInt(buildTime, 10, 64)
	if err != nil {
		return versionInfo, errors.New("failed to parse build time")
	}
	versionInfo.BuildTime = time.Unix(parsedBuildTime, 0)
	return versionInfo, nil
}

// AddCheck adds a provided checker to the health check
func (hc *HealthCheck) AddCheck(name string, checker Checker) (err error) {
	_, err = hc.AddAndGetCheck(name, checker)
	return err
}

// AddAndGetCheck adds a provided checker to the health check
// and returns the corresponding Check pointer, which maybe used for subscription
func (hc *HealthCheck) AddAndGetCheck(name string, checker Checker) (check *Check, err error) {
	check, err = NewCheck(name, checker)
	if err != nil {
		return nil, err
	}
	check.state.changeCallback = hc.healthChangeCallback

	hc.Checks = append(hc.Checks, check)

	ticker := createTicker(hc.interval, check)
	hc.tickers = append(hc.tickers, ticker)

	if hc.context != nil {
		ticker.start(hc.context, hc.tickersWaitgroup)
	}

	return check, nil
}

// Start begins each ticker, this is used to run the health checks on dependent apps
// It also starts a go-routine to check the state after the criticalErrorTimeout has expired
// until the app is fully started, to make sure the state is updated accordingly without relying on the http Handle being called
// takes argument context and should utilise contextWithCancel
// Passing a nil context will cause errors during stop/app shutdown
func (hc *HealthCheck) Start(ctx context.Context) {
	hc.context = ctx
	hc.StartTime = time.Now().UTC()
	for _, ticker := range hc.tickers {
		ticker.start(ctx, hc.tickersWaitgroup)
	}
	hc.startTracker(ctx)
}

// startTracker creates a new go-routine to keep track of the app status after the critical timeout has expired
// until all checkers have run, which is required in order to set the app status to Critical if any dependency is not healthy after the critical timeout.
// Further updates will be performed by the callback when any checker state changes (or when the http handler is called)
func (hc *HealthCheck) startTracker(ctx context.Context) {
	hc.stopper = make(chan struct{})
	hc.tickersWaitgroup.Add(1)
	go func(ctx context.Context) {
		defer hc.tickersWaitgroup.Done()
		for {
			select {
			case <-time.After(hc.criticalErrorTimeout):
				hc.loopAppStartingUp(ctx)
			case <-ctx.Done():
				return
			case <-hc.stopper:
				return
			}
		}
	}(ctx)
}

// loopAppStartingUp polls the app state until it has fully started (all checks have run)
// - The polling period is 10% of the checkers interval, with jitter.
// - The state is updated to 'WARNING' until the app has fully started: then it is updated to the real status according to checkers.
func (hc *HealthCheck) loopAppStartingUp(ctx context.Context) {
	intervalWithJitter := calcIntervalWithJitter(hc.interval / 10)
	for {
		select {
		case <-time.After(intervalWithJitter):
			if hc.isAppStartingUp() {
				log.Warn(ctx, "a dependency is still starting up")
				hc.SetStatusAndUptime(StatusWarning)
				continue
			}

			newStatus := hc.isAppHealthy()
			hc.SetStatusAndUptime(newStatus)
			return

		case <-ctx.Done():
			return
		case <-hc.stopper:
			return
		}
	}
}

// Stop will cancel all tickers and thus stop all health checks
// It also stops the go-routine that was created in start, if it is still alive.
func (hc *HealthCheck) Stop() {
	for _, ticker := range hc.tickers {
		ticker.stop()
	}
	close(hc.stopper)
	hc.tickersWaitgroup.Wait()
}

// GetStatus returns the current status in a thread-safe way
func (hc *HealthCheck) GetStatus() string {
	hc.statusLock.RLock()
	defer hc.statusLock.RUnlock()
	return hc.Status
}

// SetStatus sets the current status in a thread-safe way,
// returns the new status
func (hc *HealthCheck) SetStatus(newStatus string) string {
	hc.statusLock.Lock()
	defer hc.statusLock.Unlock()
	hc.Status = newStatus
	return newStatus
}

// SetStatusAndUptime sets the current status and update Uptime according to the current time in a thread-safe way,
// returns the new status
func (hc *HealthCheck) SetStatusAndUptime(newStatus string) string {
	hc.statusLock.Lock()
	defer hc.statusLock.Unlock()
	now := time.Now().UTC()
	hc.Status = newStatus
	hc.Uptime = now.Sub(hc.StartTime) / time.Millisecond
	return newStatus
}
