package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
)

// A list of possible health check status codes
const (
	StatusOK       = "OK"
	StatusCritical = "CRITICAL"
	StatusWarning  = "WARNING"
)

// Handler responds to an http request for the current health status
func (hc HealthCheck) Handler(w http.ResponseWriter, req *http.Request) {
	now := time.Now().UTC()
	ctx := req.Context()

	var checks []Check

	for _, client := range hc.Clients {
		if client.Check != nil {
			checks = append(checks, *client.Check)
		}
	}

	hc.Status = hc.getStatus()
	hc.Uptime = now.Sub(hc.StartTime)
	hc.Checks = checks

	b, err := json.Marshal(hc)
	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "failed to marshal json"), log.Data{"error": err, "health_check_response": hc})
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrorCtx(ctx, errors.Wrap(err, "failed to write bytes for http response"), nil)
		return
	}
}

// isAppStartingUp returns false when all clients have completed at least one check
func (hc HealthCheck) isAppStartingUp() bool {
	for _, client := range hc.Clients {
		if client.Check == nil {
			return true
		}
	}
	return false
}

// getStatus returns a status as string as to the overall current apps health based on its dependent apps health
func (hc HealthCheck) getStatus() string {
	if hc.isAppStartingUp() {
		log.Info("a dependency is still starting up", nil)
		return StatusWarning
	}
	return hc.isAppHealthy()
}

// isAppHealthy checks every clients check for their health then produces and returns a status for this apps health
func (hc HealthCheck) isAppHealthy() string {
	status := StatusOK
	for _, client := range hc.Clients {
		appHealth := hc.readAppHealth(client)
		if appHealth == StatusCritical {
			return StatusCritical
		} else if appHealth == StatusWarning {
			status = StatusWarning
		}
	}
	return status
}

// readAppHealth locks mutex then reads a check finally it unlocks the mutex.
func (hc HealthCheck) readAppHealth(client *Client) string {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	return hc.isCheckHealthy(client.Check)
}

// isCheckHealthy returns a string for the status on if an individual dependent apps health
func (hc HealthCheck) isCheckHealthy(c *Check) string {
	now := time.Now().UTC()
	switch c.Status {
	case StatusWarning:
		return StatusWarning
	case StatusOK:
		return StatusOK
	default:
		status := StatusWarning
		criticalTimeThreshold := hc.TimeOfFirstCriticalError.Add(hc.CriticalErrorTimeout)
		if c.LastSuccess.Before(hc.TimeOfFirstCriticalError) && now.After(criticalTimeThreshold) {
			status = StatusCritical
		}
		// Set timestamp of first critical error to now
		if c.LastSuccess.After(hc.TimeOfFirstCriticalError) {
			hc.TimeOfFirstCriticalError = now
		}
		return status
	}
}
