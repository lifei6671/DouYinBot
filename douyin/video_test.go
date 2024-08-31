package douyin

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestVideo_Download(t *testing.T) {
	str := `5.35 Slc:/ 谁不爱全能的小王子呢%头盔 %拍照姿势 %单眼皮  https://v.douyin.com/RtQ332e/ 複製佌链接，打开Dou音搜索，矗接观看視频！ oxBCQt9rsUybLpUJ0BqHYk1SWZR4`

	dy := NewDouYin(
		web.AppConfig.DefaultString("douyinproxy", "https://api.disign.me/api"),
		web.AppConfig.DefaultString("douyinproxyusername", ""),
		web.AppConfig.DefaultString("douyinproxypassword", ""),
	)
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
		urlStr := "http://v3-web.douyinvod.com/46dcfa120d9045b22915eef9685a83b2/66d2a43e/video/tos/cn/tos-cn-ve-15/oIA5EmZBAe2EAUF9KZEQ5AANwYgfw16S9n0IzD/?a=6383\u0026ch=26\u0026cr=3\u0026dr=0\u0026lr=all\u0026cd=0%7C0%7C0%7C3\u0026cv=1\u0026br=3285\u0026bt=3285\u0026cs=0\u0026ds=4\u0026ft=4TMWc6Dnppft2zLd.sd.C_bAja-CInniuGtc6B3U~JP2SYpHDDaPd.m-ZGgzLusZ.\u0026mime_type=video_mp4\u0026qs=0\u0026rc=NDc0O2g0NTVmZDtlZjkzaEBpM3M0Onc5cmdzdTMzNGkzM0BiMzI2Nl9iXzIxM15fMDMzYSNscDQtMmRjbzVgLS1kLWFzcw%3D%3D\u0026btag=80000e00008000\u0026cquery=100w_100B_100x_100z_100o\u0026dy_q=1725069826\u0026feature_id=46a7bb47b4fd1280f3d3825bf2b29388\u0026l=2024083110034682A8613EB0FDE26EF8C2"

		video := Video{VideoId: "a"}
		cover, err := video.DownloadCover(urlStr, "./aaa/")
		convey.So(err, convey.ShouldBeNil)
		convey.So(cover, convey.ShouldNotBeNil)

	})
}
