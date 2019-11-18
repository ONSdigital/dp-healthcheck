package healthcheck

import (
	"context"
	"errors"
	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
	"time"
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
	cfok := Checker(func(ctx *context.Context) (check *Check, err error) {
		return &healthyCheck1, nil
	})
	cfFail := Checker(func(ctx *context.Context) (*Check, error) {
		err := errors.New("checker failed to run for some app 1")
		return nil, err
	})

	Convey("Create a new Health Check given one good working check function to run", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 1 * time.Millisecond
		checkerFunc := cfok
		checkerFuncPointer := &checkerFunc
		client := Client{
			Clienter:   rchttp.NewClient(),
			Checker:    checkerFuncPointer,
			mutex: &sync.Mutex{},
		}
		clients := []*Client{&client}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(ctx, version, criticalTimeout, interval, clients)
		So(hc.Clients[0], ShouldEqual, &client)
		So(hc.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime.After(timeBeforeCreation), ShouldEqual, true)
		So(hc.StartTime.Before(time.Now().UTC()), ShouldEqual, true)
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.tickers), ShouldEqual, 1)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * time.Millisecond)
			checkResponse := *hc.tickers[0].client.Check
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
			Clienter:   rchttp.NewClient(),
			Checker:    checkerFuncPointer,
			mutex: &sync.Mutex{},
		}
		clients := []*Client{&client}
		hc := Create(ctx, version, criticalTimeout, interval, clients)
		Convey("After check function has run, ensure it has correctly stored the results", func() {
			time.Sleep(2 * time.Millisecond)
			So(hc.tickers[0].client.Check, ShouldBeNil)
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
			Clienter:   rchttp.NewClient(),
			Checker:    checkerFuncPointer,
			mutex: &sync.Mutex{},
		}
		client.Check = &healthyCheck1
		clients := []*Client{&client}
		hc := Create(ctx, version, criticalTimeout, interval, clients)
		Convey("After check function has run, the original check should not be overwritten by the failed check", func() {
			time.Sleep(2 * time.Millisecond)
			checkResponse := *hc.tickers[0].client.Check
			So(checkResponse, ShouldResemble, healthyCheck1)
		})
	})
}
