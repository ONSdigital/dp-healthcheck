package healthcheck

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"time"
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
// checkers which implement the checker interface and can run a checkup to determine the health of the app and/or its dependencies
func New(version VersionInfo, criticalTimeout, interval time.Duration, checkers ...Checker) (HealthCheck, error) {
	hc := HealthCheck{
		Checks:               []*Check{},
		Version:              version,
		criticalErrorTimeout: criticalTimeout,
		interval:             interval,
		tickers:              []*ticker{},
	}

	for _, checker := range checkers {
		if err := hc.AddCheck(checker); err != nil {
			return hc, err
		}
	}

	return hc, nil
}

// CreateVersionInfo returns a health check version info object. Caller to provide:
// buildTime for when the app was built as a unix time stamp in string form
// gitCommit the SHA-1 commit hash of the built app
// version the semantic version of the built app
func CreateVersionInfo(buildTime, gitCommit, version string) (VersionInfo, error) {
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
func (hc *HealthCheck) AddCheck(checker Checker) (err error) {
	check, err := NewCheck(checker)
	if err != nil {
		return err
	}

	hc.Checks = append(hc.Checks, check)

	ticker := createTicker(hc.interval, check)
	hc.tickers = append(hc.tickers, ticker)

	if hc.context != nil {
		ticker.start(hc.context)
	}

	return nil
}

// Start begins each ticker, this is used to run the health checks on dependent apps
// takes argument context and should utilise contextWithCancel
// Passing a nil context will cause errors during stop/app shutdown
func (hc *HealthCheck) Start(ctx context.Context) {
	hc.context = ctx
	hc.StartTime = time.Now().UTC()
	for _, ticker := range hc.tickers {
		ticker.start(ctx)
	}
}

// Stop will cancel all tickers and thus stop all health checks
func (hc *HealthCheck) Stop() {
	for _, ticker := range hc.tickers {
		ticker.stop()
	}
}
