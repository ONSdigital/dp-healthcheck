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

func createHealthCheck(statuses []CheckState, startTime time.Time, critErrTimeout time.Duration, pretendHistory bool) HealthCheck {
	hc := HealthCheck{
		Checks:               createChecksSlice(statuses, pretendHistory),
		Version:              testVersion,
		StartTime:            startTime,
		CriticalErrorTimeout: critErrTimeout,
		Tickers:              nil,
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

func runHealthHandlerAndTest(t *testing.T, hc *HealthCheck, desiredStatus string, testVersion VersionInfo, testStartTime time.Time, statuses []CheckState) {
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

func TestHandlerSignleCheck(t *testing.T) {
	t0 := time.Now().UTC()
	t10 := t0.Add(-10 * time.Minute) // 10 min ago
	t20 := t0.Add(-20 * time.Minute) // 20 min ago
	t30 := t0.Add(-30 * time.Minute) // 30 min ago
	criticalErrTimeout := 11 * time.Minute

	healthyStatus1 := CheckState{
		Name:        "Some App 1",
		Status:      StatusOK,
		StatusCode:  http.StatusOK,
		Message:     "App 1 is healthy",
		LastChecked: &t0,
		LastSuccess: &t0,
		LastFailure: &t10,
	}
	unhealthyStatus := CheckState{
		Name:        "Some App 2",
		Status:      StatusWarning,
		StatusCode:  http.StatusTooManyRequests,
		Message:     "Something has been unhealthy for past 10 minutes",
		LastChecked: &t0,
		LastSuccess: &t10,
		LastFailure: &t0,
	}
	freshCriticalStatus := CheckState{
		Name:        "Some App 3",
		Status:      StatusCritical,
		StatusCode:  http.StatusInternalServerError,
		Message:     "Something has been critical for the past 10 minutes",
		LastChecked: &t0,
		LastSuccess: &t10,
		LastFailure: &t0,
	}
	oldCriticalStatus := CheckState{
		Name:        "Some App 4",
		Status:      StatusCritical,
		StatusCode:  http.StatusInternalServerError,
		Message:     "Something has been critical for the past 30 minutes",
		LastChecked: &t0,
		LastSuccess: &t30,
		LastFailure: &t0,
	}

	nilStatus := CheckState{
		Name: "Some App 5",
	}
	healthyNeverUnhealthyStatus := CheckState{
		Name:        "Some App 6",
		Status:      StatusOK,
		StatusCode:  http.StatusOK,
		Message:     "App 6 is healthy",
		LastChecked: &t0,
		LastSuccess: &t0,
	}
	unhealthyNeverHealthyStatus := CheckState{
		Name:        "Some App 7",
		Status:      StatusWarning,
		StatusCode:  http.StatusTooManyRequests,
		Message:     "Something is unhealthy",
		LastChecked: &t0,
		LastFailure: &t0,
	}
	criticalNeverHealthyStatus := CheckState{
		Name:        "Some App 8",
		Status:      StatusCritical,
		StatusCode:  http.StatusTooManyRequests,
		Message:     "Something is critical",
		LastChecked: &t0,
		LastFailure: &t0,
	}

	Convey("Given a healthcheck with no past failures or successes", t, func() {

		hc := HealthCheck{
			Version:              testVersion,
			StartTime:            t0,
			CriticalErrorTimeout: criticalErrTimeout,
			Tickers:              nil,
		}

		Convey("An empty check should result in the app reporting back as warning", func() {
			statuses := []CheckState{nilStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not set
			So(hc.TimeOfFirstCriticalError, ShouldResemble, time.Time{})
		})
		Convey("A healthy check that has never been unhealthy should result in the app reporting back as healthy", func() {
			statuses := []CheckState{healthyNeverUnhealthyStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusOK, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not set
			So(hc.TimeOfFirstCriticalError, ShouldResemble, time.Time{})
		})
		Convey("A healthy check that has been unhealthy in the past should result in the app reporting back as healthy", func() {
			statuses := []CheckState{healthyStatus1}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusOK, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not set
			So(hc.TimeOfFirstCriticalError, ShouldResemble, time.Time{})
		})
		Convey("An unhealthy check that has never been healthy should result in the app reporting back as warning", func() {
			statuses := []CheckState{unhealthyNeverHealthyStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not set
			So(hc.TimeOfFirstCriticalError, ShouldResemble, time.Time{})
		})
		Convey("An unhealthy check that has been healthy in the past should result in the app reporting back as warning", func() {
			statuses := []CheckState{unhealthyStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not set
			So(hc.TimeOfFirstCriticalError, ShouldResemble, time.Time{})
		})
		Convey("A critical check that has never been healthy should result in the app reporting back as warning and updating timestamp for first critical error", func() {
			statuses := []CheckState{criticalNeverHealthyStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError set to this check's failure time
			So(hc.TimeOfFirstCriticalError, ShouldHappenWithin, time.Second, t0)
		})
		Convey("A critical check that has been healthy in the past should result in the app reporting back as warning and updating timestamp for first critical error", func() {
			statuses := []CheckState{oldCriticalStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError set to this check's failure time
			So(hc.TimeOfFirstCriticalError, ShouldHappenWithin, time.Second, t0)
		})
	})

	Convey("Given a healthcheck with a recent past critical check (timeout not expired), and no success received since", t, func() {

		hc := HealthCheck{
			Version:                  testVersion,
			StartTime:                t0,
			CriticalErrorTimeout:     criticalErrTimeout,
			TimeOfFirstCriticalError: t10,
			Tickers:                  nil,
		}

		Convey("A healthy check should result in the app reporting back as healthy", func() {
			statuses := []CheckState{healthyNeverUnhealthyStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusOK, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not updated
			So(hc.TimeOfFirstCriticalError, ShouldResemble, t10)
		})
		Convey("A recent critical check happening before the timeout expires should result in the app reporting back as warning", func() {
			statuses := []CheckState{freshCriticalStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not updated
			So(hc.TimeOfFirstCriticalError, ShouldResemble, t10)
		})
		Convey("A critical check that has been critical for longer than the timeout and the value of first critical error, "+
			"should result in the app reporting back as warning and not updating timestamp for first critical error", func() {
			statuses := []CheckState{oldCriticalStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not updated
			So(hc.TimeOfFirstCriticalError, ShouldResemble, t10)
		})
	})

	Convey("Given a healthcheck with an old past critical check (timeout expired), and no success received since", t, func() {

		hc := HealthCheck{
			Version:                  testVersion,
			StartTime:                t0,
			CriticalErrorTimeout:     criticalErrTimeout,
			TimeOfFirstCriticalError: t20,
			Tickers:                  nil,
		}

		Convey("A healthy check should result in the app reporting back as healthy", func() {
			statuses := []CheckState{healthyNeverUnhealthyStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusOK, testVersion, t0, statuses)
			// TimeOfFirstCriticalError not set
			So(hc.TimeOfFirstCriticalError, ShouldResemble, t20)
		})
		Convey("A recent critical check (last success more recent than first critical) should result in the app reporting back as warning "+
			"and refresh timestamp for first critical error", func() {
			statuses := []CheckState{freshCriticalStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, t0, statuses)
			// TimeOfFirstCriticalError set to this check's failure time
			So(hc.TimeOfFirstCriticalError, ShouldHappenWithin, time.Second, t0)
		})
		Convey("A critical check (last success older than first critical) should result in the app reporting back as critical "+
			"and not refreshing timestamp for first critical error", func() {
			statuses := []CheckState{oldCriticalStatus}
			hc.Checks = createChecksSlice(statuses, true)
			runHealthHandlerAndTest(t, &hc, StatusCritical, testVersion, t0, statuses)
			// TimeOfFirstCriticalError set to this check's failure time
			So(hc.TimeOfFirstCriticalError, ShouldResemble, t20)
		})
	})

}

func TestHandlerMultipleChecks(t *testing.T) {
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
		hc := createHealthCheck(statuses, testStartTime, 10*time.Minute, true)
		hc.TimeOfFirstCriticalError = testStartTime.Add(-30 * time.Minute)
		runHealthHandlerAndTest(t, &hc, StatusOK, testVersion, testStartTime, statuses)
	})
	Convey("Given a healthy app and an unhealthy app", t, func() {
		statuses := []CheckState{healthyStatus1, unhealthyStatus}
		hc := createHealthCheck(statuses, testStartTime, 15*time.Second, true)
		hc.TimeOfFirstCriticalError = testStartTime.Add(-30 * time.Minute)
		runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, testStartTime, statuses)
	})
	Convey("Given a healthy app and a critical app that is beyond the threshold", t, func() {
		checks := []CheckState{healthyStatus1, criticalStatus}
		hc := createHealthCheck(checks, testStartTime, 10*time.Minute, true)
		hc.TimeOfFirstCriticalError = testStartTime.Add(-22 * time.Minute)
		runHealthHandlerAndTest(t, &hc, StatusCritical, testVersion, testStartTime, checks)
	})
	Convey("Given an unhealthy app and an app that has just turned critical and is under the critical threshold", t, func() {
		statuses := []CheckState{unhealthyStatus, freshCriticalStatus}
		hc := createHealthCheck(statuses, testStartTime, 10*time.Minute, true)
		hc.TimeOfFirstCriticalError = testStartTime.Add(-1 * time.Minute)
		runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, testStartTime, statuses)
	})
	Convey("Given an unhealthy app and an app that has been critical for longer than the critical threshold", t, func() {
		statuses := []CheckState{unhealthyStatus, criticalStatus}
		hc := createHealthCheck(statuses, testStartTime, 10*time.Minute, true)
		hc.TimeOfFirstCriticalError = testStartTime.Add(-22 * time.Minute)
		runHealthHandlerAndTest(t, &hc, StatusCritical, testVersion, testStartTime, statuses)
	})
	Convey("Given an app just started up", t, func() {
		statuses := []CheckState{freshCriticalStatus}
		justStartedTime := time.Now().UTC()
		hc := createHealthCheck(statuses, justStartedTime, 10*time.Minute, false)
		hc.TimeOfFirstCriticalError = justStartedTime
		runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, justStartedTime, nil)
	})
	Convey("Given an app has begun to start but not finished starting up completely", t, func() {
		statuses := []CheckState{freshCriticalStatus}
		justStartedTime := time.Now().UTC()
		hc := createHealthCheck(statuses, justStartedTime, 10*time.Minute, true)
		hc.TimeOfFirstCriticalError = justStartedTime
		runHealthHandlerAndTest(t, &hc, StatusWarning, testVersion, justStartedTime, statuses)
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
		runHealthHandlerAndTest(t, &hc, StatusOK, testVersion, testStartTime, statuses)
	})
}
