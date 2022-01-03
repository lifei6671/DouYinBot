package douyin

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDouYin_Get(t *testing.T) {
	convey.Convey("DouYin_Get",t, func() {
		content := "5.87 WZZ:/ 再见少年拉弓满、不惧岁月不惧风！  https://v.douyin.com/85MyVfe/ 复制此链接，达开Douyin搜索，矗接观看视pin！ oxBCQt9rsUybLpUJ0BqHYk1SWZR4"

		dy := NewDouYin()

		convey.Convey("DouYin_Get_OK", func() {
			video,err := dy.Get(content)
			convey.So(err,convey.ShouldBeNil)
			convey.So(video,convey.ShouldNotBeNil)
		})
	})
}
