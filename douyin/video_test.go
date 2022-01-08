package douyin

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestVideo_Download(t *testing.T) {
	str := `5.35 Slc:/ 谁不爱全能的小王子呢%头盔 %拍照姿势 %单眼皮  https://v.douyin.com/RtQ332e/ 複製佌链接，打开Dou音搜索，矗接观看視频！ oxBCQt9rsUybLpUJ0BqHYk1SWZR4`

	dy := NewDouYin()
	video, err := dy.Get(str)
	if err != nil {
		t.Fatal(err)
	}
	p, err := video.Download("./video/")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p)
}

func TestVideo_DownloadCover(t *testing.T) {
	convey.Convey("Video_DownloadCove", t, func() {
		urlStr := "https://p3-sign.douyinpic.com/tos-cn-p-0015/d584b1df940f44d2a9aff73e7935a718_1641203848~tplv-dy-360p.jpeg?x-expires=1642413600&x-signature=zACK8k4JJR4eaWkdIi0CW3nSHOs%3D&from=4257465056&s=&se=false&sh=&sc=&l=202201031834420102121450193E098A1D&biz_tag=feed_cover"

		video := Video{VideoId: "a"}
		video.DownloadCover(urlStr,"/aaa/")

	})
}
