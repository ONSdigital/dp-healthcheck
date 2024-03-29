package healthcheck

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

var minTime = time.Unix(0, 0)

// Handler responds to an http request for the current health status
func (hc *HealthCheck) Handler(w http.ResponseWriter, req *http.Request) {
	hc.statusLock.Lock()
	defer hc.statusLock.Unlock()

	now := time.Now().UTC()
	ctx := req.Context()

	newStatus := hc.getAppStatus(ctx)
	hc.Status = newStatus
	hc.Uptime = now.Sub(hc.StartTime) / time.Millisecond

	b, err := json.Marshal(hc)
	if err != nil {
		log.Error(ctx, "failed to marshal json", err, log.Data{"health_check_response": hc})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch newStatus {
	case StatusOK:
		w.WriteHeader(http.StatusOK)
	case StatusWarning:
		w.WriteHeader(http.StatusTooManyRequests)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, err = w.Write(b)
	if err != nil {
		log.Error(ctx, "failed to write bytes for http response", err)
		return
	}
}

// isAppStartingUp returns false when all clients have completed at least one check
func (hc *HealthCheck) isAppStartingUp() bool {
	return hc.areChecksStartingUp(hc.Checks)
}

func (hc *HealthCheck) areChecksStartingUp(checks []*Check) bool {
	for _, check := range checks {
		if !check.hasRun() {
			return true
		}
	}
	return false
}

// getAppStatus returns a status as string as to the overall current apps health based on its dependent apps health
func (hc *HealthCheck) getAppStatus(ctx context.Context) string {
	if hc.isAppStartingUp() {
		log.Warn(ctx, "a dependency is still starting up")
		return StatusWarning
	}
	return hc.isAppHealthy()
}

func (hc *HealthCheck) getChecksStatus(checks []*Check) string {
	if hc.areChecksStartingUp(checks) {
		return StatusWarning
	}
	return hc.areChecksHealthy(checks)
}

// isAppHealthy checks individual Checks for their health then produces
// and returns the 'worst' of those statuses as this app's health
// (i.e. if any are StatusCritical, return that,
//          else if any are StatusWarning, return that,
//          else return StatusOK)
func (hc *HealthCheck) isAppHealthy() string {
	return hc.areChecksHealthy(hc.Checks)
}

func (hc *HealthCheck) areChecksHealthy(checks []*Check) string {
	status := StatusOK
	for _, check := range checks {
		checkStatus := hc.getCheckStatus(check)
		if checkStatus == StatusCritical {
			return StatusCritical
		} else if checkStatus == StatusWarning {
			status = StatusWarning
		}
	}
	return status
}

// getCheckStatus returns a string for the status on an individual check
func (hc *HealthCheck) getCheckStatus(c *Check) string {
	switch c.state.Status() {
	case StatusOK:
		return StatusOK
	case StatusWarning:
		return StatusWarning
	default:

		now := time.Now().UTC()
		status := StatusWarning

		// last success or minTime if nil. c should not be muted.
		lastSuccess := c.state.LastSuccess()
		if lastSuccess == nil {
			lastSuccess = &minTime
		}

		// Global state will be considered critical if check has been critical for longer
		// than the first critical error since last success and the timeout has expired.
		criticalTimeThreshold := hc.timeOfFirstCriticalError.Add(hc.criticalErrorTimeout)
		if lastSuccess.Before(hc.timeOfFirstCriticalError) && now.After(criticalTimeThreshold) {
			status = StatusCritical
		}

		// Set timestamp of first critical error to now if there has been a success since the previous value, or if this is the first one.
		if lastSuccess.After(hc.timeOfFirstCriticalError) || hc.timeOfFirstCriticalError.IsZero() {
			hc.timeOfFirstCriticalError = now
		}

		return status
	}
}
