package controllers

import (
	"errors"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/admin/structs"
	"github.com/lifei6671/douyinbot/internal/utils"
)

type ContentController struct {
	web.Controller
}

func (c *ContentController) Index() {
	videoId := c.Ctx.Input.Param(":video_id")

	if videoId == "" {
		c.Ctx.Output.SetStatus(404)
		return
	}
	video, err := models.NewDouYinVideo().FirstByVideoId(videoId)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		return
	}
	c.Data["desc"] = video.Desc
	html, err := models.NewDouYinTag().FormatTagHtml(video.Desc)
	if err != nil {
		logs.Error("处理视频标签失败[video_id=%s] %+v", video.VideoId, err)
	} else {
		video.Desc = html
	}
	c.Data["video"] = video

	c.TplName = "index/content.gohtml"
}

type VideoResult struct {
	VideoId       string `json:"video_id"`
	Cover         string `json:"cover"`
	PlayAddr      string `json:"play_addr"`
	LocalPlayAddr string `json:"local_play_addr"`
	AuthorURL     string `json:"author_url"`
	Nickname      string `json:"nickname"`
	Desc          string `json:"desc"`
}

func (c *ContentController) Next() {
	videoId := c.Ctx.Input.Query("video_id")
	action := c.Ctx.Input.Query("action")

	if videoId == "" || action == "" || (action != "next" && action != "prev") {
		c.Ctx.Output.SetStatus(404)
		logs.Error("请求参数异常：[video_id=%s, action=%s]", videoId, action)
		return
	}
	var video *models.DouYinVideo
	var err error
	if action == "next" {
		video, err = models.NewDouYinVideo().Next(videoId)
	} else {
		video, err = models.NewDouYinVideo().Prev(videoId)
	}
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			ret := structs.JsonResult[*VideoResult]{
				ErrCode: 404,
				Message: utils.Ternary(action == "next", "已经是最后一页啦", "已经是第一页啦"),
			}
			_ = c.JSONResp(ret)
			return
		}
		logs.Error("视频翻页出错：[video_id=%s, action=%s] %+v", videoId, action, err)
		c.Ctx.Output.SetStatus(500)
		return
	}
	ret := structs.JsonResult[*VideoResult]{
		ErrCode: 0,
		Message: "",
		Data: &VideoResult{
			VideoId:       video.VideoId,
			Cover:         video.VideoLocalCover,
			PlayAddr:      video.VideoPlayAddr,
			LocalPlayAddr: c.URLFor("VideoController.Index", "video_id", video.VideoId),
			AuthorURL:     c.URLFor("IndexController.List", ":author_id", video.AuthorId, ":page", 1),
			Nickname:      video.Nickname,
			Desc:          video.Desc,
		},
	}
	html, err := models.NewDouYinTag().FormatTagHtml(video.Desc)
	if err != nil {
		logs.Error("处理视频标签失败[video_id=%s, action=%s] %+v", video.VideoId, action, err)
	} else {
		ret.Data.Desc = html
	}

	_ = c.JSONResp(ret)
}
