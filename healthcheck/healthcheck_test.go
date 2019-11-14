package healthcheck

import (
	"context"
	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	Convey("Create a new Health Check", t, func() {
		ctx := context.Background()
		version := "1.0.0"
		criticalTimeout := 15 * time.Second
		interval := 3 * time.Second
		checkerFunc := Checker(func(ctx *context.Context) (check *Check, err error) {
			return
		})
		checkerFuncPointer := &checkerFunc
		client := Client{
			Clienter:   rchttp.NewClient(),
			Check:      nil,
			Checker:    checkerFuncPointer,
			MutexCheck: &sync.Mutex{},
		}
		clients := []*Client{&client}
		timeBeforeCreation := time.Now().UTC()
		hc := Create(ctx, version, criticalTimeout, interval, clients)
		timeAfterCreation := time.Now().UTC()
		So(hc.Clients[0], ShouldEqual, &client)
		So(hc.Version, ShouldEqual, "1.0.0")
		So(hc.StartTime.After(timeBeforeCreation), ShouldEqual, true)
		So(hc.StartTime.Before(timeAfterCreation), ShouldEqual, true)
		So(hc.CriticalErrorTimeout, ShouldEqual, criticalTimeout)
		So(len(hc.Tickers), ShouldEqual, 1)
	})
}
// TODO ... How to test ticker.Ticker stop func
func TestStop(t *testing.T) {
	Convey("Stop tickers", t, func() {

	})
}

