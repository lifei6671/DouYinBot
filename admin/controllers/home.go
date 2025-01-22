package controllers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/beego/beego/v2/server/web"

	"github.com/lifei6671/douyinbot/admin/structs"
	"github.com/lifei6671/douyinbot/douyin"
)

var douYin *douyin.DouYin
var once sync.Once

var (
	videoHtml = `<video controls="controls" autoplay="autoplay" width="100%"><source src="{{__VIDEO__}}" type="video/mp4"></video>`
)

type HomeController struct {
	web.Controller
}

func (c *HomeController) Index() {
	once.Do(func() {
		douYin = douyin.NewDouYin(
			web.AppConfig.DefaultString("douyinproxy", ""),
			web.AppConfig.DefaultString("douyinproxyusername", ""),
			web.AppConfig.DefaultString("douyinproxypassword", ""),
		)
	})
	if c.Ctx.Input.IsGet() {
		c.TplName = "home/index.gohtml"
	} else {
		douYinContent := c.Ctx.Input.Query("douYinContent")
		if douYinContent == "" {
			c.Data["json"] = &structs.JsonResult[string]{
				ErrCode: 1,
				Message: "解析内容失败",
			}
		} else {
			//service.Push(context.Background(), service.MediaContent{
			//	Content: douYinContent,
			//	UserId:  "lifei6671",
			//})
			//return
			video, err := douYin.Get(douYinContent)
			if err != nil {
				c.Data["json"] = &structs.JsonResult[string]{
					ErrCode: 1,
					Message: err.Error(),
				}
			} else if video.VideoType == douyin.VideoPlayType {
				c.Data["json"] = &structs.JsonResult[string]{
					ErrCode: 0,
					Message: "ok",
					Data:    strings.ReplaceAll(videoHtml, "{{__VIDEO__}}", video.PlayAddr),
				}
			} else if video.VideoType == douyin.ImagePlayType {
				var imageHtml string

				for _, image := range video.Images {
					imageHtml += fmt.Sprintf(`<p><a href="%s" target="_blank"><img src="%s"></a></p>`,
						image.ImageUrl, image.ImageUrl,
					)
				}
				c.Data["json"] = &structs.JsonResult[string]{
					ErrCode: 0,
					Message: "ok",
					Data:    imageHtml,
				}
			} else {
				c.Data["json"] = &structs.JsonResult[string]{
					ErrCode: 1,
					Message: "无法解析",
				}
			}

		}
		c.ServeJSON()
	}
}

func (c *HomeController) Download() {
	urlStr := c.Ctx.Input.Query("url")
	if urlStr == "" {
		c.Data["json"] = &structs.JsonResult[string]{
			ErrCode: 1,
			Message: "获取抖音地址失败",
		}
	} else {
		video, err := douYin.Get(urlStr)
		if err != nil {
			c.Data["json"] = &structs.JsonResult[string]{
				ErrCode: 1,
				Message: err.Error(),
			}
		} else {
			if c.Ctx.Input.IsAjax() {
				location, _ := video.GetDownloadUrl()

				c.Data["json"] = &structs.JsonResult[map[string]string]{
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
					c.Data["json"] = &structs.JsonResult[string]{
						ErrCode: 2,
						Message: "未找到文件保存目录",
						Data:    video.PlayAddr,
					}
				} else {
					_, err = video.Download(filename)
					if err != nil {
						c.Data["json"] = &structs.JsonResult[string]{
							ErrCode: 1,
							Message: err.Error(),
						}
					} else {
						c.Data["json"] = &structs.JsonResult[string]{
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
