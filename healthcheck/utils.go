package healthcheck

import (
	"math/rand"
	"time"
)

// calcIntervalWithJitter returns a new duration based on a provided interval and a jitter percentage of 5%
func calcIntervalWithJitter(interval time.Duration) time.Duration {
	jitterPercentage := 0.05
	jitterThreshold := int(float64(interval) * jitterPercentage)
	jitterAmount := time.Duration(random(-1*jitterThreshold, jitterThreshold))
	return interval + jitterAmount
}

// random creates a random integer between min and max
func random(min, max int) int {
	return rand.Intn(max-min) + min
}
