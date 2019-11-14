package healthcheck

import (
	"context"
	"encoding/json"
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

func createATestClient(checkToReturn Check) *Client {
	checkerFunc := createATestChecker(checkToReturn)
	cli := NewClient(nil, checkerFunc)
	cli.Check = &checkToReturn
	return cli
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
		Name:        "Some App",
		Status:      StatusWarning,
		StatusCode:  429,
		Message:     "Something has been unhealthy for past 30 minutes",
		LastChecked: testStartTime,
		LastSuccess: testStartTime.Add(time.Duration(-30) * time.Minute),
		LastFailure: testStartTime,
	}
	//
	criticalCheck := Check{
		Name:        "Some App",
		Status:      StatusCritical,
		StatusCode:  500,
		Message:     "Something has been critical for the past 30 minutes",
		LastChecked: testStartTime,
		LastSuccess: testStartTime.Add(time.Duration(-30) * time.Minute),
		LastFailure: testStartTime,
	}
	//log.Info("Just for compiling whist working", log.Data{"hmm": healthyCheck,
	//	"arg": unhealthyCheck, "grr": criticalCheck})

	Convey("Given a complete Healthy set of checks the app should report back as healthy", t, func() {
		var clients []*Client
		checks := []Check{healthyCheck1, healthyCheck2}
		for _, check := range checks {
			clients = append(clients, createATestClient(check))
		}
		// Create a HealthCheck response
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: testStartTime.Add(time.Duration(-30) * time.Minute),
			Tickers:                  nil,
		}
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
		var healthResponse HealthResponse
		err = json.Unmarshal(b, &healthResponse)
		So(w.Code, ShouldEqual, 200)
		So(healthResponse.Status, ShouldEqual, StatusOK)
		So(healthResponse.Version, ShouldEqual, testVersion)
		So(healthResponse.StartTime, ShouldEqual, testStartTime)
		So(healthResponse.Checks, ShouldResemble, checks)
		So(healthResponse.Uptime, ShouldNotBeNil)
		So(time.Now().UTC().After(healthResponse.StartTime.Add(healthResponse.Uptime)), ShouldBeTrue)
	})
	Convey("Given a healthy app and an unhealthy app", t, func() {
		var clients []*Client
		checks := []Check{healthyCheck1, unhealthyCheck}
		for _, check := range checks {
			clients = append(clients, createATestClient(check))
		}
		// Create a HealthCheck response
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     15 * time.Second,
			TimeOfFirstCriticalError: testStartTime.Add(time.Duration(-30) * time.Minute),
			Tickers:                  nil,
		}
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
		var healthResponse HealthResponse
		err = json.Unmarshal(b, &healthResponse)
		So(w.Code, ShouldEqual, 200)
		So(healthResponse.Status, ShouldEqual, StatusWarning)
		So(healthResponse.Version, ShouldEqual, testVersion)
		So(healthResponse.StartTime, ShouldEqual, testStartTime)
		So(healthResponse.Checks, ShouldResemble, checks)
		So(healthResponse.Uptime, ShouldNotBeNil)
		So(time.Now().UTC().After(healthResponse.StartTime.Add(healthResponse.Uptime)), ShouldBeTrue)
	})
	Convey("Given a healthy app and an app starting up", t, func() {
		var clients []*Client
		justStartedTime := time.Now().UTC().Add(time.Duration(-1) * time.Minute)
		checks := []Check{healthyCheck1, unhealthyCheck}
		for _, check := range checks {
			clients = append(clients, createATestClient(check))
		}
		// Create a HealthCheck response
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                justStartedTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: testStartTime.Add(time.Duration(-30) * time.Minute),
			Tickers:                  nil,
		}
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
		var healthResponse HealthResponse
		err = json.Unmarshal(b, &healthResponse)
		So(w.Code, ShouldEqual, 200)
		So(healthResponse.Status, ShouldEqual, StatusWarning)
		So(healthResponse.Version, ShouldEqual, testVersion)
		So(healthResponse.StartTime, ShouldEqual, justStartedTime)
		So(healthResponse.Checks, ShouldResemble, checks)
		So(healthResponse.Uptime, ShouldNotBeNil)
		So(time.Now().UTC().After(healthResponse.StartTime.Add(healthResponse.Uptime)), ShouldBeTrue)
	})
	Convey("Given no apps", t, func() {
		var clients []*Client
		var checks []Check
		// Create a HealthCheck response
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: testStartTime.Add(time.Duration(-30) * time.Minute),
			Tickers:                  nil,
		}
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
		var healthResponse HealthResponse
		err = json.Unmarshal(b, &healthResponse)
		So(w.Code, ShouldEqual, 200)
		So(healthResponse.Status, ShouldEqual, StatusOK)
		So(healthResponse.Version, ShouldEqual, testVersion)
		So(healthResponse.StartTime, ShouldEqual, testStartTime)
		So(healthResponse.Checks, ShouldResemble, checks)
		So(healthResponse.Uptime, ShouldNotBeNil)
		So(time.Now().UTC().After(healthResponse.StartTime.Add(healthResponse.Uptime)), ShouldBeTrue)
	})
	Convey("Given a healthy app and a critical app that is beyond the threshold", t, func() {
		var clients []*Client
		checks := []Check{healthyCheck1, criticalCheck}
		for _, check := range checks {
			clients = append(clients, createATestClient(check))
		}
		// Create a HealthCheck response
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: testStartTime.Add(time.Duration(-22) * time.Minute),
			Tickers:                  nil,
		}
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
		var healthResponse HealthResponse
		err = json.Unmarshal(b, &healthResponse)
		So(w.Code, ShouldEqual, 200)
		So(healthResponse.Status, ShouldEqual, StatusCritical)
		So(healthResponse.Version, ShouldEqual, testVersion)
		So(healthResponse.StartTime, ShouldEqual, testStartTime)
		So(healthResponse.Checks, ShouldResemble, checks)
		So(healthResponse.Uptime, ShouldNotBeNil)
		So(time.Now().UTC().After(healthResponse.StartTime.Add(healthResponse.Uptime)), ShouldBeTrue)
	})
	Convey("Given an unhealthy app and an app that has just turned critical and is under the critical threshold", t, func() {
		var clients []*Client
		checks := []Check{unhealthyCheck, criticalCheck}
		for _, check := range checks {
			clients = append(clients, createATestClient(check))
		}
		// Create a HealthCheck response
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: time.Now().Add(time.Duration(-1) * time.Minute),
			Tickers:                  nil,
		}
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
		var healthResponse HealthResponse
		err = json.Unmarshal(b, &healthResponse)
		So(w.Code, ShouldEqual, 200)
		So(healthResponse.Status, ShouldEqual, StatusWarning)
		So(healthResponse.Version, ShouldEqual, testVersion)
		So(healthResponse.StartTime, ShouldEqual, testStartTime)
		So(healthResponse.Checks, ShouldResemble, checks)
		So(healthResponse.Uptime, ShouldNotBeNil)
		So(time.Now().UTC().After(healthResponse.StartTime.Add(healthResponse.Uptime)), ShouldBeTrue)
	})
	Convey("Given an unhealthy app and an app that has been critical for longer than the critical threshold", t, func() {
		var clients []*Client
		checks := []Check{unhealthyCheck, criticalCheck}
		for _, check := range checks {
			clients = append(clients, createATestClient(check))
		}
		// Create a HealthCheck response
		hc := HealthCheck{
			Clients:                  clients,
			Version:                  testVersion,
			StartTime:                testStartTime,
			CriticalErrorTimeout:     10 * time.Minute,
			TimeOfFirstCriticalError: testStartTime.Add(time.Duration(-22) * time.Minute),
			Tickers:                  nil,
		}
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
		var healthResponse HealthResponse
		err = json.Unmarshal(b, &healthResponse)
		So(w.Code, ShouldEqual, 200)
		So(healthResponse.Status, ShouldEqual, StatusCritical)
		So(healthResponse.Version, ShouldEqual, testVersion)
		So(healthResponse.StartTime, ShouldEqual, testStartTime)
		So(healthResponse.Checks, ShouldResemble, checks)
		So(healthResponse.Uptime, ShouldNotBeNil)
		So(time.Now().UTC().After(healthResponse.StartTime.Add(healthResponse.Uptime)), ShouldBeTrue)
	})
}
