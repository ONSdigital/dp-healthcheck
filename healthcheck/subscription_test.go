package healthcheck

import (
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSubscribe(t *testing.T) {
	s1 := &mock.SubscriberMock{}
	c1 := &Check{}
	c2 := &Check{}
	c3 := &Check{}

	Convey("Given a healthcheck with an empty list of subscribers", t, func() {
		hc := &HealthCheck{
			subscribers: map[Subscriber]map[*Check]struct{}{},
			subsMutex:   &sync.Mutex{},
			Checks:      []*Check{c1, c2, c3},
		}

		Convey("Then Subscribe a subscriber results in the internal map being modified accordingly", func() {
			hc.Subscribe(s1, c1, c2)
			So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
				s1: {c1: {}, c2: {}},
			})

			Convey("Then Subscribe the same subscriber again with different checks results in the union of checks being subscribed", func() {
				hc.Subscribe(s1, c1, c3)
				So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
					s1: {c1: {}, c2: {}, c3: {}},
				})
			})
		})

		Convey("Then calling SubscribeAll results in the subscriber being added to the map with all the checkers", func() {
			hc.SubscribeAll(s1)
			So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
				s1: {c1: {}, c2: {}, c3: {}},
			})
		})
	})
}

func TestUnsubscribe(t *testing.T) {
	sub1 := &mock.SubscriberMock{}
	sub2 := &mock.SubscriberMock{}
	c1 := &Check{}
	c2 := &Check{}
	c3 := &Check{}

	Convey("Given a healthcheck with 2 subscribers and 2 checkers per subscriber", t, func() {
		hc := &HealthCheck{
			subscribers: map[Subscriber]map[*Check]struct{}{
				sub1: {c1: {}, c2: {}},
				sub2: {c2: {}, c3: {}},
			},
			subsMutex: &sync.Mutex{},
			Checks:    []*Check{c1, c2, c3},
		}

		Convey("Then Unsubscribe one subscriber results in only the expected item being removed from the internal map", func() {
			hc.Unsubscribe(sub1, c2)
			So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
				sub1: {c1: {}},
				sub2: {c2: {}, c3: {}},
			})

			Convey("Then Unsubscribing the last check results in the subscriber being removed from the external map", func() {
				hc.Unsubscribe(sub1, c1)
				So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
					sub2: {c2: {}, c3: {}},
				})
			})
		})

		Convey("Then Calling UnsubscribeAll results in the subscriber being removed from the external map", func() {
			hc.UnsubscribeAll(sub1)
			So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
				sub2: {c2: {}, c3: {}},
			})
		})

		Convey("Then Unsubscribing a subscriber that was not subscrived has no effect", func() {
			hc.Unsubscribe(&mock.SubscriberMock{}, c1, c2)
			So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
				sub1: {c1: {}, c2: {}},
				sub2: {c2: {}, c3: {}},
			})

			hc.UnsubscribeAll(&mock.SubscriberMock{})
			So(hc.subscribers, ShouldResemble, map[Subscriber]map[*Check]struct{}{
				sub1: {c1: {}, c2: {}},
				sub2: {c2: {}, c3: {}},
			})
		})
	})
}

func TestNotifyHealthUpdate(t *testing.T) {
	t0 := time.Now().UTC()
	t10 := t0.Add(-10 * time.Minute) // 10 min ago

	sub1 := &mock.SubscriberMock{OnHealthUpdateFunc: func(status string) {}}
	sub2 := &mock.SubscriberMock{OnHealthUpdateFunc: func(status string) {}}
	c1 := createATestCheck(CheckState{
		status:      StatusOK,
		lastChecked: &t0,
		lastSuccess: &t0,
		lastFailure: &t10,
	}, true)
	c2 := createATestCheck(CheckState{
		status:      StatusOK,
		lastChecked: &t0,
		lastSuccess: &t0,
		lastFailure: &t10,
	}, true)
	c3 := createATestCheck(CheckState{
		status:      StatusCritical,
		lastChecked: &t0,
		lastSuccess: &t0,
		lastFailure: &t10,
	}, true)

	Convey("Given a healthcheck with a total of 3 checks, 2 subscribers and 2 checkers per subscriber", t, func() {
		hc := &HealthCheck{
			Checks: []*Check{c1, c2, c3},
			subscribers: map[Subscriber]map[*Check]struct{}{
				sub1: {c1: {}, c2: {}},
				sub2: {c2: {}, c3: {}},
			},
			statusLock: &sync.RWMutex{},
			subsMutex:  &sync.Mutex{},
		}

		Convey(`Then calling healthChangeCallback results in the global app status and the combined status for all the subscribers being updated`, func() {
			So(hc.GetStatus(), ShouldEqual, "") // app status before callback
			wg := hc.healthChangeCallback()
			wg.Wait()
			So(sub1.OnHealthUpdateCalls(), ShouldHaveLength, 1)
			So(sub2.OnHealthUpdateCalls(), ShouldHaveLength, 1)
			So(sub1.OnHealthUpdateCalls()[0].Status, ShouldEqual, StatusOK)      // combined status of {c1, c2}
			So(sub2.OnHealthUpdateCalls()[0].Status, ShouldEqual, StatusWarning) // combined status of {c2, c3}
			So(hc.GetStatus(), ShouldEqual, StatusCritical)                      // app status after callback
		})
	})
}
