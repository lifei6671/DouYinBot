package controllers

import (
	"context"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/douyin"
	"github.com/tidwall/gjson"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	defaultVideoUrl     = "https://api.amemv.com/aweme/v1/play/?video_id=v0200f480000br2flq7iv420dp6l9js0&ratio=480p&line=1"
	defaultVideoContent []byte
	bm, _               = cache.NewCache("memory", `{"interval":60}`)
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

	video, err := models.NewDouYinVideo().FirstByVideoId(videoId)
	if err != nil {
		c.sendFile("")
		return
	}
	dir := web.AppConfig.DefaultString("auto-save-path", "")
	if dir == "" {
		c.sendFile("")
		logs.Warn("没有配置本地储存路径 -> %s", videoId)
		return
	}
	filename := filepath.Join(dir, video.VideoLocalAddr)
	c.sendFile(filename)
}

func (c *VideoController) Play() {
	videoId := c.Ctx.Input.Query("video_id")
	if videoId == "" {
		c.Ctx.Abort(404, "param err")
		return
	}
	cid, err := bm.Get(context.Background(), videoId)
	if err == nil {
		c.Ctx.Redirect(301, cid.(string))
		c.StopRun()
	}

	video, err := models.NewDouYinVideo().FirstByVideoId(videoId)
	if err != nil {
		c.Ctx.Abort(404, "param err")
		return
	}
	if len(video.AwemeId) == 0 {
		c.Ctx.Abort(404, "")
		return
	}
	dy := douyin.NewDouYin(
		web.AppConfig.DefaultString("douyinproxy", ""),
		web.AppConfig.DefaultString("douyinproxyusername", ""),
		web.AppConfig.DefaultString("douyinproxypassword", ""),
	)
	b, err := dy.GetVideoInfo(video.VideoRawPlayAddr)
	if err != nil {
		logs.Error(err)
		c.Ctx.Abort(500, "get video failed")
	}
	playURL := gjson.Get(b, "video_data.wm_video_url_HQ").String()
	if len(playURL) > 0 {
		_ = bm.Put(context.Background(), videoId, playURL, time.Hour*24)
	}
	c.Ctx.Redirect(301, playURL)
	c.StopRun()
}

func (c *VideoController) sendFile(filename string) {
	if stat, err := os.Stat(filename); os.IsNotExist(err) || stat.IsDir() {
		logs.Warn("文件不存在 -> %s", filename)
		if defaultVideoContent == nil || len(defaultVideoContent) == 0 {
			c.Redirect(defaultVideoUrl, http.StatusFound)
		} else {
			c.Ctx.Output.Header("Content-Type", "video/mp4")
			_ = c.Ctx.Output.Body(defaultVideoContent)
		}
		return
	}

	c.Ctx.Output.Header("Content-Type", "video/mp4")
	http.ServeFile(c.Ctx.ResponseWriter, c.Ctx.Request, filename)
	c.StopRun()
}
func SetDefaultVideoContent(body []byte) {
	defaultVideoContent = body
}
