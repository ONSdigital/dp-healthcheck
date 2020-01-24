package healthcheck

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/log.go/log"
	. "github.com/smartystreets/goconvey/convey"
)

var testVersion = VersionInfo{
	BuildTime:       time.Unix(0, 0),
	GitCommit:       "d6cd1e2bd19e03a81132a23b2025920577f84e37",
	Language:        "go",
	LanguageVersion: "1.12",
	Version:         "1.0.0",
}

func createATestCheck(stateToReturn CheckState, pretendHistory bool) *Check {
	checkerFunc := func(ctx context.Context, state *CheckState) error {
		if pretendHistory {
			state = &stateToReturn
		}
		return nil
	}
	check, _ := newCheck(checkerFunc)
	if pretendHistory {
		check.state = &stateToReturn
	}
	return check
}

func createHealthCheck(statuses []CheckState, startTime time.Time, critErrTimeout time.Duration, firstCritErr time.Time, pretendHistory bool) HealthCheck {
	hc := HealthCheck{
		Checks:                   createChecksSlice(statuses, pretendHistory),
		Version:                  testVersion,
		StartTime:                startTime,
		CriticalErrorTimeout:     critErrTimeout,
		TimeOfFirstCriticalError: firstCritErr,
		Tickers:                  nil,
	}
	return hc
}

func createChecksSlice(statuses []CheckState, pretendHistory bool) []*Check {
	var checks []*Check
	for _, status := range statuses {
		checks = append(checks, createATestCheck(status, pretendHistory))
	}
	return checks
}

func runHealthHandlerAndTest(t *testing.T, hc HealthCheck, desiredStatus string, testVersion VersionInfo, testStartTime time.Time, statuses []CheckState) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fail()
	}
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(hc.Handler)
	handler.ServeHTTP(w, req)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		log.Event(nil, "unable to read request body", log.Error(err))
	}
	var healthCheck HealthCheck
	err = json.Unmarshal(b, &healthCheck)
	if err != nil {
		log.Event(nil, "unable to unmarshal bytes into healthcheck", log.Error(err))

		So(err, ShouldBeNil)
		return
	}
	So(w.Code, ShouldEqual, http.StatusOK)
	So(healthCheck.Status, ShouldEqual, desiredStatus)
	So(healthCheck.Version, ShouldResemble, testVersion)
	So(healthCheck.StartTime, ShouldEqual, testStartTime)
	So(healthCheck.Uptime, ShouldNotBeNil)
	So(time.Now().UTC().After(healthCheck.StartTime.Add(healthCheck.Uptime)), ShouldBeTrue)

	if statuses != nil {
		for i, check := range healthCheck.Checks {
			So(*check.state, ShouldResemble, statuses[i])
		}
	} else {

	}
}

func TestHandler(t *testing.T) {
	testStartTime := time.Now().UTC().Add(-20 * time.Minute)
	priorTestTime := testStartTime.Add(-30 * time.Minute)
	healthyStatus1 := CheckState{
		Name:        "Some App 1",
		Status:      StatusOK,
		StatusCode:  http.StatusOK,
		Message:     "Some message about app 1 here",
		LastChecked: &testStartTime,
		LastSuccess: &testStartTime,
		LastFailure: &priorTestTime,
	}
	healthyStatus2 := CheckState{
		Name:        "Some App 2",
		Status:      StatusOK,
		StatusCode:  http.StatusOK,
		Message:     "Some message about app 2 here",
		LastChecked: &testStartTime,
		LastSuccess: &testStartTime,
		LastFailure: &priorTestTime,
	}
	healthyStatus3 := CheckState{
		Name:        "Some App 3",
		Status:      StatusOK,
		Message:     "Some message about app 2 here",
		LastChecked: &testStartTime,
		LastSuccess: &testStartTime,
		LastFailure: &priorTestTime,
	}
	unhealthyStatus := CheckState{
		Name:        "Some App 4",
		Status:      StatusWarning,
		StatusCode:  http.StatusTooManyRequests,
		Message:     "Something has been unhealthy for past 30 minutes",
		LastChecked: &testStartTime,
		LastSuccess: &priorTestTime,
		LastFailure: &testStartTime,
	}
	criticalStatus := CheckState{
		Name:        "Some App 5",
		Status:      StatusCritical,
		StatusCode:  http.StatusInternalServerError,
		Message:     "Something has been critical for the past 30 minutes",
		LastChecked: &testStartTime,
		LastSuccess: &priorTestTime,
		LastFailure: &testStartTime,
	}
	freshCriticalStatus := CheckState{
		Name:        "Some App 6",
		Status:      StatusCritical,
		StatusCode:  http.StatusInternalServerError,
		Message:     "Something has been critical for the past 30 minutes",
		LastChecked: &testStartTime,
		LastSuccess: &testStartTime,
		LastFailure: &priorTestTime,
	}

	Convey("Given a complete Healthy set of checks the app should report back as healthy", t, func() {
		statuses := []CheckState{healthyStatus1, healthyStatus2, healthyStatus3}
		hc := createHealthCheck(statuses, testStartTime, 10*time.Minute, testStartTime.Add(-30*time.Minute), true)
		runHealthHandlerAndTest(t, hc, StatusOK, testVersion, testStartTime, statuses)
	})
	Convey("Given a healthy app and an unhealthy app", t, func() {
		statuses := []CheckState{healthyStatus1, unhealthyStatus}
		hc := createHealthCheck(statuses, testStartTime, 15*time.Second, testStartTime.Add(-30*time.Minute), true)
		runHealthHandlerAndTest(t, hc, StatusWarning, testVersion, testStartTime, statuses)
	})
	Convey("Given a healthy app and a critical app that is beyond the threshold", t, func() {
		checks := []CheckState{healthyStatus1, criticalStatus}
		hc := createHealthCheck(checks, testStartTime, 10*time.Minute, testStartTime.Add(-22*time.Minute), true)
		runHealthHandlerAndTest(t, hc, StatusCritical, testVersion, testStartTime, checks)
	})
	Convey("Given an unhealthy app and an app that has just turned critical and is under the critical threshold", t, func() {
		statuses := []CheckState{unhealthyStatus, freshCriticalStatus}
		hc := createHealthCheck(statuses, testStartTime, 10*time.Minute, time.Now().Add(-1*time.Minute), true)
		runHealthHandlerAndTest(t, hc, StatusWarning, testVersion, testStartTime, statuses)
	})
	Convey("Given an unhealthy app and an app that has been critical for longer than the critical threshold", t, func() {
		statuses := []CheckState{unhealthyStatus, criticalStatus}
		hc := createHealthCheck(statuses, testStartTime, 10*time.Minute, testStartTime.Add(-22*time.Minute), true)
		runHealthHandlerAndTest(t, hc, StatusCritical, testVersion, testStartTime, statuses)
	})
	Convey("Given an app just started up", t, func() {
		statuses := []CheckState{freshCriticalStatus}
		justStartedTime := time.Now().UTC()
		hc := createHealthCheck(statuses, justStartedTime, 10*time.Minute, justStartedTime, false)
		runHealthHandlerAndTest(t, hc, StatusWarning, testVersion, justStartedTime, nil)
	})
	Convey("Given an app has begun to start but not finished starting up completely", t, func() {
		statuses := []CheckState{freshCriticalStatus}
		justStartedTime := time.Now().UTC()
		hc := createHealthCheck(statuses, justStartedTime, 10*time.Minute, justStartedTime, true)
		runHealthHandlerAndTest(t, hc, StatusWarning, testVersion, justStartedTime, statuses)
	})
	Convey("Given no apps", t, func() {
		var checks []*Check
		var statuses []CheckState
		hc := HealthCheck{
			Checks:                   checks,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: testStartTime.Add(-30 * time.Minute),
			Tickers:                  nil,
		}
		runHealthHandlerAndTest(t, hc, StatusOK, testVersion, testStartTime, statuses)
	})
}
