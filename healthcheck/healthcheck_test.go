package healthcheck

import (
	"context"
	"errors"
	"testing"
	"time"

	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	criticalTimeout = 15 * time.Second
	interval        = 100 * time.Millisecond
)

var version = VersionObj{
	BuildTime:       time.Unix(0, 0),
	GitCommit:       "d6cd1e2bd19e03a81132a23b2025920577f84e37",
	Language:        "go",
	LanguageVersion: "1.12",
	Version:         "1.0.0",
}

func getTestCheck(msg string) *Check {
	timeAfterCreation := time.Now().UTC()
	previousFailure := timeAfterCreation.Add(time.Duration(-30) * time.Minute)
	return &Check{
		Status:      StatusOK,
		StatusCode:  200,
		Message:     msg,
		LastChecked: &timeAfterCreation,
		LastSuccess: &timeAfterCreation,
		LastFailure: &previousFailure,
	}
}

func getTestClient(name string, checker *Checker, chk *Check) *Client {
	client, err := NewClient(
		rchttp.NewClient(),
		checker,
	)
	if err != nil {
		return nil
	}
	client.Check = chk
	return client
}

func TestCreate(t *testing.T) {
	healthyCheck1 := getTestCheck("Success from app 1")
	healthyCheck2 := getTestCheck("Success from app 2")
	healthyCheck3 := getTestCheck("Success from app 3")

	cfok1 := Checker(func(ctx context.Context) (*Check, error) {
		return healthyCheck1, nil
	})
	cfok2 := Checker(func(ctx context.Context) (*Check, error) {
		return healthyCheck2, nil
	})
	cfok3 := Checker(func(ctx context.Context) (*Check, error) {
		return healthyCheck3, nil
	})

	cfFail := Checker(func(ctx context.Context) (*Check, error) {
		err := errors.New("checker failed to run for cfFail")
		return nil, err
	})

	Convey("Create a new Health Check given one good working check function to run with status code", t, func() {
		ctx := context.Background()
		clients := []*Client{getTestClient(
			"ok1 client",
			&cfok1,
			nil,
		)}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(ctx)
		defer hc.Stop()

		hc.Tickers[0].client.mutex.Lock()
		So(hc.Clients[0], ShouldPointTo, clients[0])
		hc.Tickers[0].client.mutex.Unlock()
		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, "go")
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)
			hc.Tickers[0].client.mutex.Lock()
			checkResponse := hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})

	Convey("Create a new Health Check given one good working check function to run (with status code)", t, func() {
		ctx := context.Background()
		clients := []*Client{getTestClient(
			"ok3 check",
			&cfok3,
			nil,
		)}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(ctx)
		defer hc.Stop()

		hc.Tickers[0].client.mutex.Lock()
		So(hc.Clients[0], ShouldPointTo, clients[0])
		hc.Tickers[0].client.mutex.Unlock()

		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, "go")
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)

			hc.Tickers[0].client.mutex.Lock()
			checkResponse := hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck3)
		})
	})

	Convey("Create a new Health Check given one good working check function to run (without status code)", t, func() {
		ctx := context.Background()
		clients := []*Client{getTestClient(
			"ok1 no status",
			&cfok1,
			nil,
		)}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(ctx)
		defer hc.Stop()

		hc.Tickers[0].client.mutex.Lock()
		So(hc.Clients[0], ShouldPointTo, clients[0])
		hc.Tickers[0].client.mutex.Unlock()

		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, "go")
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)

		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)

			hc.Tickers[0].client.mutex.Lock()
			checkResponse := hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})
	Convey("Create a new Health Check given a broken check function", t, func() {
		ctx := context.Background()
		clients := []*Client{getTestClient(
			"broken check func",
			&cfFail,
			nil,
		)}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(ctx)
		defer hc.Stop()

		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)
			So(hc.Tickers[0].client.Check, ShouldBeNil)
		})
	})

	Convey("Given a Health Check with a cancellable context", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		clients := []*Client{getTestClient(
			"cancel app",
			&cfFail,
			getTestCheck("cancellable testing"),
		)}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(ctx)
		// no `defer hc.Stop()` because of `cancel()`

		So(hc.Started, ShouldBeTrue)

		Convey("When the check function has time to run, and the context is cancelled", func() {
			time.Sleep(2 * interval)

			So(len(hc.Tickers), ShouldEqual, 1)
			So(hc.Tickers[0].client.Check, ShouldPointTo, clients[0].Check)
			So(hc.Tickers[0].isStopping(), ShouldBeFalse)

			cancel()

			Convey("Then the tickers are stopped/stopping", func() {
				time.Sleep(2 * interval)
				So(hc.Tickers[0].isStopping(), ShouldBeTrue)
			})
		})
	})

	Convey("Create a new Health Check given 1 successful check followed by a broken run check", t, func() {
		ctx := context.Background()
		clients := []*Client{getTestClient(
			"test broken2",
			&cfFail,
			healthyCheck1,
		)}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(ctx)
		defer hc.Stop()

		Convey("After check function has run, the original check should not be overwritten by the failed check", func() {
			time.Sleep(2 * interval)

			hc.Tickers[0].client.mutex.Lock()
			checkResponse := hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})

	Convey("Create a new Health Check given 1 client at creation and a second added before start is called", t, func() {
		ctx := context.Background()
		clients := []*Client{getTestClient(
			"start after 2nd",
			&cfok1,
			healthyCheck1,
		)}

		hc := Create(version, criticalTimeout, interval, clients)
		err := hc.AddClient(getTestClient(
			"2nd client",
			&cfok2,
			nil,
		))
		So(err, ShouldBeNil)

		hc.Start(ctx)
		defer hc.Stop()

		Convey("After adding the second client there should be two timers on start", func() {
			time.Sleep(2 * interval)
			So(len(hc.Tickers), ShouldEqual, 2)
		})
	})

	Convey("Given a Health Check with 1 client that is started", t, func() {
		clients := []*Client{getTestClient(
			"before late client",
			&cfok1,
			healthyCheck1,
		)}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(context.Background())
		defer hc.Stop()

		origNumberOfTickers := len(hc.Tickers)
		Convey("When you add another client - too late", func() {
			err := hc.AddClient(getTestClient(
				"late client",
				&cfok2,
				healthyCheck2,
			))
			Convey("Then there should be no increase in the number of tickers", func() {
				So(err, ShouldNotBeNil)
				time.Sleep(2 * interval)
				So(len(hc.Tickers), ShouldEqual, origNumberOfTickers)
			})
		})
	})
}
