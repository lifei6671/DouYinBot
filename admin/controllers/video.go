package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"net/http"
	"os"
	"path/filepath"
)
var(
	defaultVideoUrl = "https://api.amemv.com/aweme/v1/play/?video_id=v0200f480000br2flq7iv420dp6l9js0&ratio=480p&line=1"
	defaultVideoContent []byte
)
type VideoController struct {
	web.Controller
}

func (c *VideoController) Index() {
	videoId := c.Ctx.Input.Query("video_id")
	if videoId == "" {
		c.sendFile("")
		return
	}

	video,err := models.NewDouYinVideo().FirstByVideoId(videoId)
	if err != nil {
		c.sendFile("")
		return
	}
	dir := web.AppConfig.DefaultString("auto-save-path","")
	if dir == "" {
		c.sendFile("")
		logs.Warn("没有配置本地储存路径 -> %s",videoId)
		return
	}
	filename := filepath.Join(dir,video.VideoLocalAddr)
	c.sendFile(filename)
}

func (c *VideoController) sendFile(filename string) {
	if _,err := os.Stat(filename); os.IsNotExist(err) {
		logs.Warn("文件不存在 -> %s",filename)
		if defaultVideoContent == nil || len(defaultVideoContent) == 0 {
			c.Redirect(defaultVideoUrl,http.StatusFound)
		} else {
			c.Ctx.Output.Header("Content-Type","video/mp4")
			_ = c.Ctx.Output.Body(defaultVideoContent)
		}
		return
	}

	c.Ctx.Output.Header("Content-Type","video/mp4")
	http.ServeFile(c.Ctx.ResponseWriter,c.Ctx.Request,filename)
	c.StopRun()
}
func SetDefaultVideoContent(body []byte) {
	defaultVideoContent = body
}