package douyin

import "testing"

func TestVideo_Download(t *testing.T) {
	str := `5.35 Slc:/ 谁不爱全能的小王子呢%头盔 %拍照姿势 %单眼皮  https://v.douyin.com/RtQ332e/ 複製佌链接，打开Dou音搜索，矗接观看視频！ oxBCQt9rsUybLpUJ0BqHYk1SWZR4`

	dy := NewDouYin()
	video,err := dy.Get(str)
	if err != nil {
		t.Fatal(err)
	}
	p,err := video.Download("./video/")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p)
}
