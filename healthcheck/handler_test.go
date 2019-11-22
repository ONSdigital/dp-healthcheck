package healthcheck

import (
	"context"
	"encoding/json"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createATestChecker(checkToReturn Check) *Checker {
	checkerFunc := Checker(func(ctx *context.Context) (check *Check, err error) {
		return &checkToReturn, nil
	})
	return &checkerFunc
}

func createATestClient(checkToReturn Check, pretendHistory bool) *Client {
	checkerFunc := createATestChecker(checkToReturn)
	clienter := rchttp.NewClient()
	cli, _ := NewClient(clienter, checkerFunc)
	if pretendHistory {
		cli.Check = &checkToReturn
	}
	return cli
}

func createHealthCheck(checks []Check, startTime time.Time, critErrTimeout time.Duration, firstCritErr time.Time, pretendHistory bool) HealthCheck {
	testVersion := "1.0.0"
	hc := HealthCheck{
		Clients:                  createClientsSlice(checks, pretendHistory),
		Version:                  testVersion,
		StartTime:                startTime,
		CriticalErrorTimeout:     critErrTimeout,
		TimeOfFirstCriticalError: firstCritErr,
		tickers:                  nil,
	}
	return hc
}

func createClientsSlice(checks []Check, pretendHistory bool) []*Client {
	var clients []*Client
	for _, check := range checks {
		clients = append(clients, createATestClient(check, pretendHistory))
	}
	return clients
}

func runHealthCheckHandlerAndTest(t *testing.T, hc HealthCheck, desiredStatus, testVersion string, testStartTime time.Time, checks []Check) {
	req, err := http.NewRequest("GET", "/healthcheck", nil)
	if err != nil {
		t.Fail()
	}
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(hc.Handler)
	handler.ServeHTTP(w, req)
	b, err := ioutil.ReadAll(w.Body)
	if err != nil {
		log.Error(err, nil)
	}
	var healthCheck HealthCheck
	err = json.Unmarshal(b, &healthCheck)
	if err != nil {
		log.Error(err, nil)
	}
	So(w.Code, ShouldEqual, 200)
	So(healthCheck.Status, ShouldEqual, desiredStatus)
	So(healthCheck.Version, ShouldEqual, testVersion)
	So(healthCheck.StartTime, ShouldEqual, testStartTime)
	So(healthCheck.Checks, ShouldResemble, checks)
	So(healthCheck.Uptime, ShouldNotBeNil)
	So(time.Now().UTC().After(healthCheck.StartTime.Add(healthCheck.Uptime)), ShouldBeTrue)
}

func TestHandler(t *testing.T) {
	testVersion := "1.0.0"
	testStartTime := time.Now().UTC().Add(time.Duration(-20) * time.Minute)
	healthyCheck1 := Check{
		Name:        "Some App 1",
		Status:      StatusOK,
		StatusCode:  200,
		Message:     "Some message about app 1 here",
		LastChecked: testStartTime,
		LastSuccess: testStartTime,
		LastFailure: testStartTime.Add(time.Duration(-30) * time.Minute),
	}
	healthyCheck2 := Check{
		Name:        "Some App 2",
		Status:      StatusOK,
		StatusCode:  200,
		Message:     "Some message about app 2 here",
		LastChecked: testStartTime,
		LastSuccess: testStartTime,
		LastFailure: testStartTime.Add(time.Duration(-30) * time.Minute),
	}
	unhealthyCheck := Check{
		Name:        "Some App 3",
		Status:      StatusWarning,
		StatusCode:  429,
		Message:     "Something has been unhealthy for past 30 minutes",
		LastChecked: testStartTime,
		LastSuccess: testStartTime.Add(time.Duration(-30) * time.Minute),
		LastFailure: testStartTime,
	}
	//
	criticalCheck := Check{
		Name:        "Some App 4",
		Status:      StatusCritical,
		StatusCode:  500,
		Message:     "Something has been critical for the past 30 minutes",
		LastChecked: testStartTime,
		LastSuccess: testStartTime.Add(time.Duration(-30) * time.Minute),
		LastFailure: testStartTime,
	}
	freshCriticalCheck := Check{
		Name:        "Some App 5",
		Status:      StatusCritical,
		StatusCode:  500,
		Message:     "Something has been critical for the past 30 minutes",
		LastChecked: testStartTime,
		LastSuccess: testStartTime,
		LastFailure: testStartTime.Add(time.Duration(-30) * time.Minute),
	}

	Convey("Given a complete Healthy set of checks the app should report back as healthy", t, func() {
		checks := []Check{healthyCheck1, healthyCheck2}
		hc := createHealthCheck(checks, testStartTime, 10*time.Minute, testStartTime.Add(time.Duration(-30)*time.Minute), true)
		runHealthCheckHandlerAndTest(t, hc, StatusOK, testVersion, testStartTime, checks)
	})
	Convey("Given a healthy app and an unhealthy app", t, func() {
		checks := []Check{healthyCheck1, unhealthyCheck}
		hc := createHealthCheck(checks, testStartTime, 15*time.Second, testStartTime.Add(time.Duration(-30)*time.Minute), true)
		runHealthCheckHandlerAndTest(t, hc, StatusWarning, testVersion, testStartTime, checks)
	})
	Convey("Given a healthy app and a critical app that is beyond the threshold", t, func() {
		checks := []Check{healthyCheck1, criticalCheck}
		hc := createHealthCheck(checks, testStartTime, 10*time.Minute, testStartTime.Add(time.Duration(-22)*time.Minute), true)
		runHealthCheckHandlerAndTest(t, hc, StatusCritical, testVersion, testStartTime, checks)
	})
	Convey("Given an unhealthy app and an app that has just turned critical and is under the critical threshold", t, func() {
		checks := []Check{unhealthyCheck, freshCriticalCheck}
		hc := createHealthCheck(checks, testStartTime, 10*time.Minute, time.Now().Add(time.Duration(-1)*time.Minute), true)
		runHealthCheckHandlerAndTest(t, hc, StatusWarning, testVersion, testStartTime, checks)
	})
	Convey("Given an unhealthy app and an app that has been critical for longer than the critical threshold", t, func() {
		checks := []Check{unhealthyCheck, criticalCheck}
		hc := createHealthCheck(checks, testStartTime, 10*time.Minute, testStartTime.Add(time.Duration(-22)*time.Minute), true)
		runHealthCheckHandlerAndTest(t, hc, StatusCritical, testVersion, testStartTime, checks)
	})
	Convey("Given an app just started up", t, func() {
		checks := []Check{freshCriticalCheck}
		justStartedTime := time.Now().UTC()
		hc := createHealthCheck(checks, justStartedTime, 10*time.Minute, justStartedTime, false)
		runHealthCheckHandlerAndTest(t, hc, StatusWarning, testVersion, justStartedTime, nil)
	})
	Convey("Given an app has begun to start but not finished starting up completely", t, func() {
		checks := []Check{freshCriticalCheck}
		justStartedTime := time.Now().UTC()
		hc := createHealthCheck(checks, justStartedTime, 10*time.Minute, justStartedTime, true)
		runHealthCheckHandlerAndTest(t, hc, StatusWarning, testVersion, justStartedTime, checks)
	})
	Convey("Given no apps", t, func() {
		var clients []*Client
		var checks []Check
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: testStartTime.Add(time.Duration(-30) * time.Minute),
			tickers:                  nil,
		}
		runHealthCheckHandlerAndTest(t, hc, StatusOK, testVersion, testStartTime, checks)
	})
}
