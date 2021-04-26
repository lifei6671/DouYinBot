package controllers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/structs"
	"github.com/lifei6671/douyinbot/douyin"
	"strings"
)

var douYin = douyin.NewDouYin()
var (
	videoHtml = `<video controls="controls" autoplay="autoplay" width="100%"><source src="{{__VIDEO__}}" type="video/mp4"></video>`
)

type HomeController struct {
	web.Controller
}

func (c *HomeController) Index() {
	if c.Ctx.Input.IsGet() {
		c.TplName = "home/index.gohtml"
	} else {
		douYinContent := c.Ctx.Input.Query("douYinContent")
		if douYinContent == "" {
			c.Data["json"] = &structs.JsonResult{
				ErrCode: 1,
				Message: "解析内容失败",
				Data:    douYinContent,
			}
		} else {
			video, err := douYin.Get(douYinContent)
			if err != nil {
				c.Data["json"] = &structs.JsonResult{
					ErrCode: 1,
					Message: err.Error(),
				}
			} else {
				c.Data["json"] = &structs.JsonResult{
					ErrCode: 0,
					Message: "ok",
					Data:    strings.ReplaceAll(videoHtml, "{{__VIDEO__}}", video.PlayAddr),
				}
			}
		}
		c.ServeJSON()
	}
}

func (c *HomeController) Download() {
	urlStr := c.Ctx.Input.Query("url")
	if urlStr == "" {
		c.Data["json"] = &structs.JsonResult{
			ErrCode: 1,
			Message: "获取抖音地址失败",
		}
	} else {
		video, err := douYin.Get(urlStr)
		if err != nil {
			c.Data["json"] = &structs.JsonResult{
				ErrCode: 1,
				Message: err.Error(),
			}
		} else {
			if c.Ctx.Input.IsAjax() {
				location, _ := video.GetDownloadUrl()

				c.Data["json"] = &structs.JsonResult{
					ErrCode: 0,
					Message: "ok",
					Data: map[string]string{
						"url":  location,
						"name": video.GetFilename(),
					},
				}

			} else {
				filename, err := web.AppConfig.String("auto-save-path")
				if err != nil {
					c.Data["json"] = &structs.JsonResult{
						ErrCode: 2,
						Message: "未找到文件保存目录",
						Data:    video.PlayAddr,
					}
				} else {
					err = video.Download(filename)
					if err != nil {
						c.Data["json"] = &structs.JsonResult{
							ErrCode: 1,
							Message: err.Error(),
						}
					} else {
						c.Data["json"] = &structs.JsonResult{
							ErrCode: 0,
							Message: "ok",
							Data:    video.PlayAddr,
						}
					}
				}
			}
		}
	}
	c.ServeJSON()
}
