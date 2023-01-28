package douyin

import (
	"github.com/smartystreets/goconvey/convey"
	"log"
	"testing"
)

func TestDouYin_Get(t *testing.T) {
	convey.Convey("DouYin_Get", t, func() {
		content := "5.87 WZZ:/ 再见少年拉弓满、不惧岁月不惧风！  https://v.douyin.com/85MyVfe/ 复制此链接，达开Douyin搜索，矗接观看视pin！ oxBCQt9rsUybLpUJ0BqHYk1SWZR4"

		content = "{9.25 Xzg:/ 复制打开抖音，看看【第十八年冬.的作品】“我从来不信什么天道，只信我自己”# 台词 # 好... https://v.douyin.com/BeSveAc/ oxBCQt9rsUybLpUJ0BqHYk1SWZR4}"
		dy := NewDouYin()

		convey.Convey("DouYin_Get_OK", func() {
			video, err := dy.Get(content)
			convey.So(err, convey.ShouldBeNil)
			convey.So(video, convey.ShouldNotBeNil)
		})
	})
}

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	m.Run()
}
