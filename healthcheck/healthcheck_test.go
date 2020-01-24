package healthcheck

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	criticalTimeout = 15 * time.Second
	interval        = 100 * time.Millisecond
)

var version = VersionInfo{
	BuildTime:       time.Unix(0, 0),
	GitCommit:       "d6cd1e2bd19e03a81132a23b2025920577f84e37",
	Language:        language,
	LanguageVersion: "1.12",
	Version:         "1.0.0",
}

func getTestCheck(msg string) *CheckState {
	timeAfterCreation := time.Now().UTC()
	previousFailure := timeAfterCreation.Add(time.Duration(-30) * time.Minute)
	return &CheckState{
		Status:      StatusOK,
		StatusCode:  200,
		Message:     msg,
		LastChecked: &timeAfterCreation,
		LastSuccess: &timeAfterCreation,
		LastFailure: &previousFailure,
	}
}

func TestCreate(t *testing.T) {
	healthyCheck1 := getTestCheck("Success from app 1")
	healthyCheck2 := getTestCheck("Success from app 2")
	healthyCheck3 := getTestCheck("Success from app 3")

	cfok1 := func(ctx context.Context) (*CheckState, error) {
		return healthyCheck1, nil
	}
	cfok2 := func(ctx context.Context) (*CheckState, error) {
		return healthyCheck2, nil
	}
	cfok3 := func(ctx context.Context) (*CheckState, error) {
		return healthyCheck3, nil
	}

	cfFail := func(ctx context.Context) (*CheckState, error) {
		err := errors.New("checker failed to run for cfFail")
		return &CheckState{}, err
	}

	Convey("Create a new Health Check given one good working check function to run with status code", t, func() {
		ctx := context.Background()
		timeBeforeCreation := time.Now().UTC()
		hc, err := Create(version, criticalTimeout, interval, cfok1)
		hc.Start(ctx)
		defer hc.Stop()

		hc.Tickers[0].check.mutex.Lock()
		So(hc.Checks[0].Checker, ShouldEqual, cfok1)
		hc.Tickers[0].check.mutex.Unlock()

		So(err, ShouldBeNil)
		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, language)
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)
			hc.Tickers[0].check.mutex.Lock()
			checkResponse := hc.Tickers[0].check.State
			hc.Tickers[0].check.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})

	Convey("Create a new Health Check given one good working check function to run (with status code)", t, func() {
		ctx := context.Background()
		timeBeforeCreation := time.Now().UTC()
		hc, err := Create(version, criticalTimeout, interval, cfok3)
		hc.Start(ctx)
		defer hc.Stop()

		hc.Tickers[0].check.mutex.Lock()
		So(hc.Checks[0].Checker, ShouldEqual, cfok3)
		hc.Tickers[0].check.mutex.Unlock()

		So(err, ShouldBeNil)
		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, language)
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)

			hc.Tickers[0].check.mutex.Lock()
			checkResponse := hc.Tickers[0].check.State
			hc.Tickers[0].check.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck3)
		})
	})

	Convey("Create a new Health Check given one good working check function to run (without status code)", t, func() {
		ctx := context.Background()
		timeBeforeCreation := time.Now().UTC()
		hc, err := Create(version, criticalTimeout, interval, cfok1)
		hc.Start(ctx)
		defer hc.Stop()

		hc.Tickers[0].check.mutex.Lock()
		So(hc.Checks[0].Checker, ShouldEqual, cfok1)
		hc.Tickers[0].check.mutex.Unlock()

		So(err, ShouldBeNil)
		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, language)
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)

		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)

			hc.Tickers[0].check.mutex.Lock()
			checkResponse := hc.Tickers[0].check.State
			hc.Tickers[0].check.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})

	Convey("Create a new Health Check given two good working check functions to run (with status code)", t, func() {
		ctx := context.Background()
		timeBeforeCreation := time.Now().UTC()
		hc, err := Create(version, criticalTimeout, interval, cfok2, cfok3)
		hc.Start(ctx)
		defer hc.Stop()

		hc.Tickers[0].check.mutex.Lock()
		So(hc.Checks[0].Checker, ShouldEqual, cfok2)
		hc.Tickers[0].check.mutex.Unlock()

		hc.Tickers[1].check.mutex.Lock()
		So(hc.Checks[1].Checker, ShouldEqual, cfok3)
		hc.Tickers[1].check.mutex.Unlock()

		So(err, ShouldBeNil)
		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, language)
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 2)
		Convey("After check functions have run, ensure they have correctly stored the results", func() {
			time.Sleep(2 * interval)

			hc.Tickers[0].check.mutex.Lock()
			checkResponse1 := hc.Tickers[0].check.State
			hc.Tickers[0].check.mutex.Unlock()
			So(checkResponse1, ShouldResemble, healthyCheck2)

			hc.Tickers[1].check.mutex.Lock()
			checkResponse2 := hc.Tickers[1].check.State
			hc.Tickers[1].check.mutex.Unlock()
			So(checkResponse2, ShouldResemble, healthyCheck3)
		})
	})

	Convey("Create a new Health Check without giving any check functions", t, func() {
		ctx := context.Background()
		timeBeforeCreation := time.Now().UTC()
		hc, err := Create(version, criticalTimeout, interval)
		hc.Start(ctx)
		defer hc.Stop()

		So(err, ShouldBeNil)
		So(hc.Version.BuildTime, ShouldEqual, time.Unix(0, 0))
		So(hc.Version.GitCommit, ShouldEqual, "d6cd1e2bd19e03a81132a23b2025920577f84e37")
		So(hc.Version.Language, ShouldEqual, language)
		So(hc.Version.LanguageVersion, ShouldEqual, "1.12")
		So(hc.Version.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenBetween, timeBeforeCreation, time.Now().UTC())
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Checks), ShouldEqual, 0)
		So(len(hc.Tickers), ShouldEqual, 0)
	})

	Convey("Create a new Health Check given a broken check function", t, func() {
		ctx := context.Background()
		hc, err := Create(version, criticalTimeout, interval, cfFail)
		hc.Start(ctx)
		defer hc.Stop()

		So(err, ShouldBeNil)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * interval)
			So(hc.Tickers[0].check.State, ShouldBeNil)
		})
	})

	Convey("Fail to create a new Health Check when given a nil check function", t, func() {
		_, err := Create(version, criticalTimeout, interval, nil)

		So(err, ShouldNotBeNil)
	})

	Convey("Given a Health Check with a cancellable context", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		hc, err := Create(version, criticalTimeout, interval, cfFail)
		hc.Checks[0].State = getTestCheck("cancellable testing")
		hc.Start(ctx)
		// no `defer hc.Stop()` because of `cancel()`

		So(err, ShouldBeNil)
		So(hc.Started, ShouldBeTrue)

		Convey("When the check function has time to run, and the context is cancelled", func() {
			time.Sleep(2 * interval)

			So(len(hc.Tickers), ShouldEqual, 1)
			So(hc.Tickers[0].check.State, ShouldPointTo, hc.Checks[0].State)
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
		hc, err := Create(version, criticalTimeout, interval, cfFail)
		hc.Checks[0].State = healthyCheck1
		hc.Start(ctx)
		defer hc.Stop()

		So(err, ShouldBeNil)

		Convey("After check function has run, the original check should not be overwritten by the failed check", func() {
			time.Sleep(2 * interval)

			hc.Tickers[0].check.mutex.Lock()
			checkResponse := hc.Tickers[0].check.State
			hc.Tickers[0].check.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})
}

func TestAddCheck(t *testing.T) {
	cf := func(ctx context.Context) (*CheckState, error) {
		return &CheckState{}, nil
	}

	Convey("Given a Health Check without any registered checks", t, func() {
		ctx := context.Background()
		hc, err := Create(version, criticalTimeout, interval)

		So(err, ShouldBeNil)

		Convey("After adding a check there should be one timer on start", func() {
			err := hc.AddCheck(cf)
			So(err, ShouldBeNil)

			hc.Start(ctx)
			defer hc.Stop()

			time.Sleep(2 * interval)
			So(len(hc.Tickers), ShouldEqual, 1)
		})
	})

	Convey("Given a Health Check with 1 check registered at creation", t, func() {
		ctx := context.Background()
		hc, err := Create(version, criticalTimeout, interval, cf)

		So(err, ShouldBeNil)

		Convey("After adding the second check there should be two timers on start", func() {
			err := hc.AddCheck(cf)
			So(err, ShouldBeNil)

			hc.Start(ctx)
			defer hc.Stop()

			time.Sleep(2 * interval)
			So(len(hc.Tickers), ShouldEqual, 2)
		})
	})

	Convey("Given a Health Check with 1 check that is started", t, func() {
		hc, err := Create(version, criticalTimeout, interval, cf)
		hc.Start(context.Background())
		defer hc.Stop()

		So(err, ShouldBeNil)
		origNumberOfTickers := len(hc.Tickers)
		Convey("When you add another check - too late", func() {
			err := hc.AddCheck(cf)
			Convey("Then there should be no increase in the number of tickers", func() {
				So(err, ShouldNotBeNil)
				time.Sleep(2 * interval)
				So(len(hc.Tickers), ShouldEqual, origNumberOfTickers)
			})
		})
	})

	Convey("Given a Health Check without any registered checks", t, func() {
		ctx := context.Background()
		hc, err := Create(version, criticalTimeout, interval)

		So(err, ShouldBeNil)

		Convey("Then adding a check with a nil checker function should fail", func() {
			err := hc.AddCheck(nil)
			So(err, ShouldNotBeNil)

			hc.Start(ctx)
			defer hc.Stop()

			time.Sleep(2 * interval)
			So(len(hc.Tickers), ShouldEqual, 0)
		})
	})
}

func TestCreateVersionInfo(t *testing.T) {
	Convey("Create a new versionInfo object", t, func() {
		buildTime := "0"
		gitCommit := "d6cd1e2bd19e03a81132a23b2025920577f84e37"
		version := "1.0.0"

		expectedVersion := VersionInfo{
			BuildTime:       time.Unix(0, 0),
			GitCommit:       gitCommit,
			Language:        language,
			LanguageVersion: runtime.Version(),
			Version:         version,
		}

		outputVersion, err := CreateVersionInfo(buildTime, gitCommit, version)

		So(err, ShouldBeNil)
		So(outputVersion, ShouldResemble, expectedVersion)
	})

	Convey("Create a new versionInfo object passing an invalid build time", t, func() {
		buildTime := "some invalid date"
		gitCommit := "d6cd1e2bd19e03a81132a23b2025920577f84e37"
		version := "1.0.0"

		expectedVersion := VersionInfo{
			BuildTime:       time.Unix(0, 0),
			GitCommit:       gitCommit,
			Language:        language,
			LanguageVersion: runtime.Version(),
			Version:         version,
		}

		outputVersion, err := CreateVersionInfo(buildTime, gitCommit, version)

		So(err, ShouldNotBeNil)
		So(outputVersion, ShouldResemble, expectedVersion)
	})
}
