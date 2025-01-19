package controllers

import (
	"math"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"github.com/lifei6671/douyinbot/admin/models"
)

type IndexController struct {
	web.Controller
}

func (c *IndexController) Index() {
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

	c.TplName = "index/index.gohtml"
}

func (c *IndexController) List() {
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
	if len(list) > 0 {
		c.Data["Nickname"] = list[0].Nickname
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

	c.TplName = "index/list.gohtml"
}
