package utils

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseExpireUnix(t *testing.T) {
	convey.Convey("ParseExpireUnix", t, func() {
		s := "https://p3-sign.douyinpic.com/obj/tos-cn-p-0015/08ad41080a594352930d76032b60cd9c_1641094018?x-expires=1642388400&x-signature=HZxZ2GJrHt58xxzWg%2Fk%2BEovCDaU%3D&from=4257465056_large"

		n, err := ParseExpireUnix(s)

		convey.So(err, convey.ShouldBeNil)
		convey.So(n, convey.ShouldEqual, 1642388400)
	})
}
