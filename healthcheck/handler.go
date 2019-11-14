package healthcheck

import (
	"encoding/json"
	"github.com/ONSdigital/go-ns/log"
	"net/http"
	"time"
)

const (
	StatusOK  = "OK"
	StatusCritical = "CRITICAL"
	StatusWarning  = "WARNING"
)

// Handler returns the current health check of the app
func (hc HealthCheck) Handler(w http.ResponseWriter, req *http.Request) {
	now := time.Now().UTC()
	ctx := req.Context()

	var checks []Check
	var status string

	status = hc.getAppHealth()
	for _, client := range hc.Clients {
		if client.Check != nil{
			checks = append(checks, *client.Check)
		}
	}

	hr := HealthResponse{
		Status:    status,
		Version:   hc.Version,
		Uptime:    now.Sub(hc.StartTime),
		StartTime: hc.StartTime,
		Checks:    checks,
	}

	b, err := json.Marshal(hr)

	if err != nil {
		log.InfoCtx(ctx, "failed to marshal json", log.Data{"error": err, "HealthCheck": hr})
		return
	}

	w.Write(b)

}

// isAppStartingUp returns true or false based on if each client has had results back from at least a single check
func (hc HealthCheck) isAppStartingUp() bool{
	for _, client := range hc.Clients {
		if client.Check == nil {
			return true
		}
	}
	return false
}

// getAppHealth returns a status as string as to the overall current apps health based on its dependent apps health
func (hc HealthCheck) getAppHealth() string {
	if hc.isAppStartingUp(){
		return StatusWarning
	}
	return hc.isOverallAppHealthy()
}

// isOverallAppHealthy checks every clients check for their health then produces and returns a status for this apps health
func (hc HealthCheck)isOverallAppHealthy() string {
	status := StatusOK
	for _, client := range hc.Clients {
		client.MutexCheck.Lock()
		appHealth := hc.isCheckHealthy(client.Check)
		client.MutexCheck.Unlock()
		if appHealth == StatusCritical{
			return StatusCritical
		} else if appHealth == StatusWarning {
			status = StatusWarning
		}
	}
	return status
}

// isCheckHealthy returns a string for the status on if an individual dependent apps health
func (hc HealthCheck)isCheckHealthy(c *Check) string {
	now := time.Now().UTC()
	status := StatusOK
	if c.Status == StatusCritical{
		criticalTimeThreshold := hc.TimeOfFirstCriticalError.Add(hc.CriticalErrorTimeout)
		if c.LastSuccess.Before(hc.TimeOfFirstCriticalError) && now.After(criticalTimeThreshold){
			status = StatusCritical
		} else {
			status = StatusWarning
		}
		// Set timestamp of first critical error to now
		if c.LastSuccess.After(hc.TimeOfFirstCriticalError) {
			hc.TimeOfFirstCriticalError = now
		}
	} else if c.Status == StatusWarning {
		status = StatusWarning
	}
	return status
}
