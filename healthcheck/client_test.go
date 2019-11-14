package healthcheck

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateNew(t *testing.T) {
	Convey("Create a new Client", t, func() {
		checkerFunc := Checker(func(ctx *context.Context) (check *Check, err error) {
			return
		})
		checkerFuncPointer := &checkerFunc

		cli := NewClient(nil, checkerFuncPointer)
		So(cli.Checker, ShouldEqual, &checkerFunc)
		So(cli.MutexCheck, ShouldNotBeNil)
		So(cli.Clienter, ShouldNotBeNil)
		So(cli.Check, ShouldBeNil)
	})

}
