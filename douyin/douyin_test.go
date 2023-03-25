package douyin

import (
	"github.com/smartystreets/goconvey/convey"
	"github.com/tidwall/gjson"
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

func TestDouYin_XBogus(t *testing.T) {
	convey.Convey("DouYin_XBogus", t, func() {
		dy := NewDouYin()

		convey.Convey("DouYin_XBogus_OK", func() {
			param := &XBogusParam{
				AwemeURL:  "https://www.douyin.com/aweme/v1/web/aweme/detail",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
			}

			sign := dy.XBogus(param)

			convey.So(sign, convey.ShouldNotBeEmpty)
			log.Println(sign)
		})

	})
}

func TestMain(m *testing.M) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	m.Run()
}

func TestDouYin_GetVideoInfo(t *testing.T) {
	convey.Convey("", t, func() {
		dy := NewDouYin()
		content := "{9.25 Xzg:/ 复制打开抖音，看看【第十八年冬.的作品】“我从来不信什么天道，只信我自己”# 台词 # 好... https://v.douyin.com/BeSveAc/ oxBCQt9rsUybLpUJ0BqHYk1SWZR4}"
		convey.Convey("", func() {
			urlStr := dy.pattern.FindString(content)
			convey.So(urlStr, convey.ShouldNotBeEmpty)

			videoId, err := dy.parseShareUrl(urlStr)
			convey.So(err, convey.ShouldBeNil)
			log.Println(videoId)

			rawUrlStr, err := dy.GetDetailUrlByVideoId(videoId)
			convey.So(err, convey.ShouldBeNil)
			b, err := dy.GetVideoInfo(rawUrlStr)
			convey.So(err, convey.ShouldBeNil)
			playURL := gjson.Get(b, "aweme_detail.video.play_addr.url_list.0").String()
			convey.So(playURL, convey.ShouldNotBeEmpty)
		})
	})
}
