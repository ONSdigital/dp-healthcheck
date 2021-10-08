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
			subscribers: map[Subscriber][]*Check{},
			subsMutex:   &sync.Mutex{},
		}

		Convey("Then Subscribe a subscriber results in the internal map being modified accordingly", func() {
			hc.Subscribe(s1, c1, c2)
			So(hc.subscribers, ShouldResemble, map[Subscriber][]*Check{
				s1: {c1, c2},
			})

			Convey("Then Subscribe the same subscriber again with different checks results in the subscriber map being updated to only the new checks", func() {
				hc.Subscribe(s1, c1, c3)
				So(hc.subscribers, ShouldResemble, map[Subscriber][]*Check{
					s1: {c1, c3},
				})
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
			subscribers: map[Subscriber][]*Check{
				sub1: {c1, c2},
				sub2: {c2, c3},
			},
			subsMutex: &sync.Mutex{},
		}

		Convey("Then Unsubscribe one subscriber results in only the expected item being removed from the internal map", func() {
			hc.Unsubscribe(sub1)
			So(hc.subscribers, ShouldResemble, map[Subscriber][]*Check{
				sub2: {c2, c3},
			})

			Convey("Then Unsubscribing the same subscriber again does not have any further effect", func() {
				hc.Unsubscribe(sub1)
				So(hc.subscribers, ShouldResemble, map[Subscriber][]*Check{
					sub2: {c2, c3},
				})
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

	Convey("Given a healthcheck with 2 subscribers and 2 checkers per subscriber", t, func() {
		hc := &HealthCheck{
			subscribers: map[Subscriber][]*Check{
				sub1: {c1, c2},
				sub2: {c2, c3},
			},
			subsMutex: &sync.Mutex{},
		}

		Convey(`Then calling healthChangeCallback results in the status being updated for all the subscribers, 
		according to the global health status of the subscribed checks for each subscriber`, func() {
			wg := hc.healthChangeCallback()
			wg.Wait()
			So(sub1.OnHealthUpdateCalls(), ShouldHaveLength, 1)
			So(sub1.OnHealthUpdateCalls()[0].Status, ShouldEqual, StatusOK)
			So(sub2.OnHealthUpdateCalls(), ShouldHaveLength, 1)
			So(sub2.OnHealthUpdateCalls()[0].Status, ShouldEqual, StatusWarning)
		})
	})
}
