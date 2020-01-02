package healthcheck

import (
	"context"
	"testing"

	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateNew(t *testing.T) {
	var (
		checkerFunc = Checker(func(ctx context.Context) (check *Check, err error) {
			return
		})
		cli      *Client
		err      error
		clienter = rchttp.NewClient()
	)
	Convey("Create a new Client", t, func() {
		cli, err = NewClient(clienter, &checkerFunc)
		So(err, ShouldBeNil)
		So(cli.Checker, ShouldPointTo, &checkerFunc)
		So(cli.mutex, ShouldNotBeNil)
		So(cli.Clienter, ShouldNotBeNil)
		So(cli.Check, ShouldBeNil)
	})
	Convey("A second new client shares the right values", t, func() {
		cli2, err := NewClient(clienter, &checkerFunc)
		So(err, ShouldBeNil)
		So(cli2.Checker, ShouldPointTo, cli.Checker)
		So(cli2.mutex, ShouldNotPointTo, cli.mutex)
	})
	Convey("Fail to create a new client as clienter given is nil", t, func() {
		cli3, err := NewClient(nil, &checkerFunc)
		So(cli3, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

}
