package healthcheck

import (
	"math/rand"
	"time"
)

const ()

// calcIntervalWithJitter returns a new duration based on a provided interval and a jitter percentage of 5%
func calcIntervalWithJitter(interval time.Duration) time.Duration {
	const jitterAmount = 0.05
	upperJitterThreshold := int(float64(interval) * jitterAmount)
	lowerJitterThreshold := -1 * upperJitterThreshold
	jitterToApply := time.Duration(random(lowerJitterThreshold, upperJitterThreshold))
	return interval + jitterToApply
}

// random creates a random integer between min and max
func random(min, max int) int {
	return rand.Intn(max-min) + min
}
