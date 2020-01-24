package healthcheck

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateNew(t *testing.T) {
	var (
		checkerFunc = func(ctx context.Context, check *CheckState) error {
			return nil
		}
		check *Check
		err   error
	)
	Convey("Create a new check", t, func() {
		check, _ = newCheck(checkerFunc)
		So(err, ShouldBeNil)
		So(check.checker, ShouldEqual, checkerFunc)
		So(check.mutex, ShouldNotBeNil)
		So(check.state, ShouldResemble, &CheckState{})
	})
	Convey("A second new check shares the right values", t, func() {
		check2, err := newCheck(checkerFunc)
		So(err, ShouldBeNil)
		So(check2.checker, ShouldEqual, check.checker)
		So(check2.mutex, ShouldNotPointTo, check.mutex)
	})
	Convey("Fail to create a new check as checker given is nil", t, func() {
		check3, err := newCheck(nil)
		So(check3, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}
