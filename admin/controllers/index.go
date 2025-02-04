package controllers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/admin/service"
	"github.com/lifei6671/douyinbot/internal/utils"
)

type IndexController struct {
	web.Controller
}

func (c *IndexController) Index() {
	if err := utils.IfLastModified(c.Ctx.Input, time.Now()); err == nil {
		c.Abort(strconv.Itoa(http.StatusNotModified))
		return
	}
	page := c.Ctx.Input.Param(":page")
	pageIndex := 1
	if page != "" {
		if num, err := strconv.Atoi(page); err == nil {
			if num <= 0 {
				c.Abort("404")
				return
			}
			pageIndex = int(math.Max(float64(num), float64(pageIndex)))
		}
	}

	list, total, err := models.NewDouYinVideo().GetList(pageIndex, 0)
	if err != nil {
		logs.Error("获取数据列表失败 -> +%+v", err)
	}
	for i, video := range list {
		if desc, err := models.NewDouYinTag().FormatTagHtml(video.Desc); err == nil {
			video.Desc = desc
			list[i] = video
		} else {
			logs.Error("渲染标签失败 ->%d - %+v", video.Id, err)
		}
		if strings.HasPrefix(video.VideoLocalCover, "/cover") {
			service.PushDownloadQueue(video)
		}
	}
	c.Data["List"] = list
	totalPage := int(math.Ceil(float64(total) / float64(models.PageSize)))

	if pageIndex <= 1 {
		c.Data["Previous"] = "#"
		c.Data["First"] = "#"
	} else {
		c.Data["Previous"] = c.URLFor("IndexController.Index", ":page", pageIndex-1)
		c.Data["First"] = c.URLFor("IndexController.Index", ":page", 1)
	}
	if pageIndex >= totalPage {
		c.Data["Next"] = "#"
		c.Data["Last"] = "#"
	} else {
		c.Data["Next"] = c.URLFor("IndexController.Index", ":page", pageIndex+1)
		c.Data["Last"] = c.URLFor("IndexController.Index", ":page", totalPage)
	}
	utils.CacheHeader(c.Ctx.Output, time.Now(), 1440, 7200)

	c.TplName = "index/index.gohtml"
}

func (c *IndexController) List() {
	if err := utils.IfLastModified(c.Ctx.Input, time.Now()); err == nil {
		c.Abort(strconv.Itoa(http.StatusNotModified))
		return
	}
	page := c.Ctx.Input.Param(":page")
	pageIndex := 1
	if page != "" {
		if num, err := strconv.Atoi(page); err == nil {
			if num <= 0 {
				c.Abort("404")
				return
			}
			pageIndex = int(math.Max(float64(num), float64(pageIndex)))
		}
	}
	authorIdStr := c.Ctx.Input.Param(":author_id")
	authorId := 0

	if authorIdStr != "" {
		if num, err := strconv.Atoi(authorIdStr); err == nil {
			authorId = num
		}
	}
	if authorId <= 0 {
		c.Abort("404")
		return
	}

	list, total, err := models.NewDouYinVideo().GetList(pageIndex, authorId)
	if err != nil {
		logs.Error("获取数据列表失败 -> +%+v", err)
	}

	if user, err := models.NewDouYinUser().GetById(authorIdStr); err == nil {
		c.Data["Desc"] = user.Signature
		if user.AvatarCDNURL != "" {
			c.Data["AvatarURL"] = user.AvatarCDNURL
		} else {
			c.Data["AvatarURL"] = user.AvatarLarger
		}

	}

	if len(list) > 0 {
		if _, ok := c.Data["NickName"]; !ok {
			c.Data["Nickname"] = list[0].Nickname
		}

		for i, video := range list {
			if desc, err := models.NewDouYinTag().FormatTagHtml(video.Desc); err == nil {
				video.Desc = desc
				list[i] = video
			} else {
				logs.Error("渲染标签失败 ->%d - %+v", video.Id, err)
			}
		}
	}
	c.Data["List"] = list
	totalPage := int(math.Ceil(float64(total) / float64(models.PageSize)))

	if pageIndex <= 1 {
		c.Data["Previous"] = "#"
		c.Data["First"] = "#"
	} else {
		c.Data["Previous"] = c.URLFor("IndexController.List", ":author_id", authorId, ":page", pageIndex-1)
		c.Data["First"] = c.URLFor("IndexController.List", ":author_id", authorId, ":page", 1)
	}
	if pageIndex >= totalPage {
		c.Data["Next"] = "#"
		c.Data["Last"] = "#"
	} else {
		c.Data["Next"] = c.URLFor("IndexController.List", ":author_id", authorId, ":page", pageIndex+1)
		c.Data["Last"] = c.URLFor("IndexController.List", ":author_id", authorId, ":page", totalPage)
	}
	utils.CacheHeader(c.Ctx.Output, time.Now(), 1440, 7200)
	c.TplName = "index/list.gohtml"
}
