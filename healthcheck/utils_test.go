package healthcheck

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJitter(t *testing.T) {

	interval := 120 * time.Second
	// this should match the calculation for upperJitterThreshold
	jitterMax := time.Duration(int64(float64(interval) * jitterFactor))
	timeRef := time.Now().UTC().Round(time.Hour)
	timeRefWithInterval := timeRef.Add(interval)

	Convey("check calcIntervalWithJitter is returning values in the expected range", t, func() {
		for i := 1; i < 20; i++ {
			timeWithJitteredInterval := timeRef.Add(calcIntervalWithJitter(interval))
			So(timeWithJitteredInterval, ShouldHappenWithin, jitterMax, timeRefWithInterval)
		}
	})
}
