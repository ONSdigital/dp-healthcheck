package healthcheck

import (
	"context"
	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateNew(t *testing.T) {
	Convey("Create a new Client", t, func() {
		checkerFunc := Checker(func(ctx *context.Context) (check *Check, err error) {
			return
		})
		checkerFuncPointer := &checkerFunc
		clienter := rchttp.NewClient()
		cli, _ := NewClient(clienter, checkerFuncPointer)
		So(cli.Checker, ShouldEqual, &checkerFunc)
		So(cli.mutex, ShouldNotBeNil)
		So(cli.Clienter, ShouldNotBeNil)
		So(cli.Check, ShouldBeNil)
	})

}
