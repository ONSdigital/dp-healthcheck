package healthcheck

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreate(t *testing.T) {
	timeAfterCreation := time.Now().UTC()
	healthyCheck1 := Check{
		Name:        "Some App 1",
		Status:      StatusOK,
		StatusCode:  200,
		Message:     "Some message about app 1 here",
		LastChecked: timeAfterCreation,
		LastSuccess: timeAfterCreation,
		LastFailure: timeAfterCreation.Add(time.Duration(-30) * time.Minute),
	}

	healthyCheck2 := Check{
		Name:        "Some App 2",
		Status:      StatusOK,
		StatusCode:  200,
		Message:     "Some message about app 2 here",
		LastChecked: timeAfterCreation,
		LastSuccess: timeAfterCreation,
		LastFailure: timeAfterCreation.Add(time.Duration(-30) * time.Minute),
	}

	healthyCheck3 := Check{
		Name:        "Some App 3",
		Status:      StatusOK,
		Message:     "Some message about app 3 here",
		LastChecked: timeAfterCreation,
		LastSuccess: timeAfterCreation,
		LastFailure: timeAfterCreation.Add(time.Duration(-30) * time.Minute),
	}

	cfok1 := Checker(func(ctx *context.Context) (check *Check, err error) {
		return &healthyCheck1, nil
	})
	cfok2 := Checker(func(ctx *context.Context) (check *Check, err error) {
		return &healthyCheck2, nil
	})
	cfok3 := Checker(func(ctx *context.Context) (check *Check, err error) {
		return &healthyCheck3, nil
	})

	cfFail := Checker(func(ctx *context.Context) (*Check, error) {
		err := errors.New("checker failed to run for some app 1")
		return nil, err
	})

	Convey("Create a new Health Check given one good working check function to run with status code", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfok1
		client := Client{
			Clienter: rchttp.NewClient(),
			Checker:  &checkerFunc,
			mutex:    &sync.Mutex{},
		}
		clients := []*Client{&client}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(&ctx)
		So(hc.Clients[0], ShouldEqual, &client)
		So(hc.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenAfter, timeBeforeCreation)
		So(hc.StartTime, ShouldHappenBefore, time.Now().UTC())
		So(hc.StartTime.Before(time.Now().UTC()), ShouldEqual, true)
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * time.Millisecond)
			hc.Tickers[0].client.mutex.Lock()
			checkResponse := *hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})

	Convey("Create a new Health Check given one good working check function to run (with status code)", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfok3
		client := Client{
			Clienter: rchttp.NewClient(),
			Checker:  &checkerFunc,
			mutex:    &sync.Mutex{},
		}
		clients := []*Client{&client}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(&ctx)
		So(hc.Clients[0], ShouldEqual, &client)
		So(hc.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenAfter, timeBeforeCreation)
		So(hc.StartTime, ShouldHappenBefore, time.Now().UTC())
		So(hc.StartTime.Before(time.Now().UTC()), ShouldEqual, true)
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * time.Millisecond)
			hc.Tickers[0].client.mutex.Lock()
			checkResponse := *hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck3)
		})
	})
	Convey("Create a new Health Check given one good working check function to run (without status code)", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfok1
		client := Client{
			Clienter: rchttp.NewClient(),
			Checker:  &checkerFunc,
			mutex:    &sync.Mutex{},
		}
		clients := []*Client{&client}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(&ctx)
		So(hc.Clients[0], ShouldEqual, &client)
		So(hc.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime, ShouldHappenAfter, timeBeforeCreation)
		So(hc.StartTime, ShouldHappenBefore, time.Now().UTC())
		So(hc.StartTime.Before(time.Now().UTC()), ShouldEqual, true)
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * time.Millisecond)
			hc.Tickers[0].client.mutex.Lock()
			checkResponse := *hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})
	Convey("Create a new Health Check given a broken check function", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfFail
		checkerFuncPointer := &checkerFunc
		client := Client{
			Clienter: rchttp.NewClient(),
			Checker:  checkerFuncPointer,
			mutex:    &sync.Mutex{},
		}
		clients := []*Client{&client}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(&ctx)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * time.Millisecond)
			So(hc.Tickers[0].client.Check, ShouldBeNil)
		})
	})
	Convey("Create a new Health Check given 1 successful check followed by a broken run check", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfFail
		checkerFuncPointer := &checkerFunc
		client := Client{
			Clienter: rchttp.NewClient(),
			Checker:  checkerFuncPointer,
			mutex:    &sync.Mutex{},
		}
		client.Check = &healthyCheck1
		clients := []*Client{&client}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(&ctx)
		Convey("After check function has run, the original check should not be overwritten by the failed check", func() {
			time.Sleep(2 * time.Millisecond)
			hc.Tickers[0].client.mutex.Lock()
			checkResponse := *hc.Tickers[0].client.Check
			hc.Tickers[0].client.mutex.Unlock()
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})

	Convey("Create a new Health Check given 1 client at creation and a second added before start is called", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfok1
		checkerFunc2 := cfok2
		checkerFuncPointer := &checkerFunc
		checkerFuncPointer2 := &checkerFunc2
		client1 := Client{
			Clienter: rchttp.NewClient(),
			Checker:  checkerFuncPointer,
			mutex:    &sync.Mutex{},
		}
		client2 := Client{
			Clienter: rchttp.NewClient(),
			Checker:  checkerFuncPointer2,
			mutex:    &sync.Mutex{},
		}
		client1.Check = &healthyCheck1
		clients := []*Client{&client1}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.AddClient(&client2)
		hc.Start(&ctx)
		Convey("After adding the second client there should be two timers on start", func() {
			time.Sleep(2 * time.Millisecond)
			So(len(hc.Tickers), ShouldEqual, 2)
		})
	})

	Convey("Create a new Health Check given 1 client at creation and a second added after start is called", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfok1
		checkerFunc2 := cfok2
		checkerFuncPointer := &checkerFunc
		checkerFuncPointer2 := &checkerFunc2
		client1 := Client{
			Clienter: rchttp.NewClient(),
			Checker:  checkerFuncPointer,
			mutex:    &sync.Mutex{},
		}
		client2 := Client{
			Clienter: rchttp.NewClient(),
			Checker:  checkerFuncPointer2,
			mutex:    &sync.Mutex{},
		}
		client1.Check = &healthyCheck1
		clients := []*Client{&client1}
		hc := Create(version, criticalTimeout, interval, clients)
		hc.Start(&ctx)
		hc.AddClient(&client2)
		Convey("After adding the second client after start there should be only 1 timer", func() {
			time.Sleep(2 * time.Millisecond)
			So(len(hc.Tickers), ShouldEqual, 1)
		})
	})
}
