package healthcheck

import (
	"context"
	"testing"

	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateNew(t *testing.T) {
	checkerFunc := Checker(func(ctx *context.Context) (check *Check, err error) {
		return
	})
	Convey("Create a new Client", t, func() {
		checkerFuncPointer := &checkerFunc
		clienter := rchttp.NewClient()
		cli, _ := NewClient(clienter, checkerFuncPointer)
		So(cli.Checker, ShouldEqual, &checkerFunc)
		So(cli.mutex, ShouldNotBeNil)
		So(cli.Clienter, ShouldNotBeNil)
		So(cli.Check, ShouldBeNil)
	})
	Convey("Fail to create a new client as clienter given is nil", t, func() {
		checkerFuncPointer := &checkerFunc
		cli, err := NewClient(nil, checkerFuncPointer)
		So(cli, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

}
