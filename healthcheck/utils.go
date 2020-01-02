package healthcheck

import (
	"math/rand"
	"time"
)

const jitterFactor = 0.05

// init seeds rand at app startup
func init() {
	rand.Seed(time.Now().UnixNano())
}

// calcIntervalWithJitter returns a new duration based on a provided interval and a jitter of ±jitterFactor
func calcIntervalWithJitter(interval time.Duration) time.Duration {
	upperJitterThreshold := int64(float64(interval) * jitterFactor)
	jitterToApply := time.Duration(random(-upperJitterThreshold, upperJitterThreshold))
	return interval + jitterToApply
}

// random creates a random integer between min and max
func random(min, max int64) int64 {
	return min + rand.Int63n(max-min)
}
