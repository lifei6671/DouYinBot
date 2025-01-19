package controllers

import (
	"math"
	"strconv"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"

	"github.com/lifei6671/douyinbot/admin/models"
)

type TagController struct {
	web.Controller
}

func (c *TagController) Index() {
	page := c.Ctx.Input.Param(":page")
	tagID := c.Ctx.Input.Param(":tag_id")
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
	if tagID == "" {
		c.Abort("404")
		return
	}
	list, tagName, total, err := models.NewDouYinTag().GetList(pageIndex, tagID)
	if err != nil {
		logs.Error("获取数据列表失败 -> +%+v", err)
	}

	if len(list) > 0 {
		c.Data["Nickname"] = tagName
	}
	c.Data["List"] = list
	totalPage := int(math.Ceil(float64(total) / float64(models.PageSize)))

	if pageIndex <= 1 {
		c.Data["Previous"] = "#"
		c.Data["First"] = "#"
	} else {
		c.Data["Previous"] = c.URLFor("TagController.List", ":tag_id", tagID, ":page", pageIndex-1)
		c.Data["First"] = c.URLFor("TagController.List", ":tag_id", tagID, ":page", 1)
	}
	if pageIndex >= totalPage {
		c.Data["Next"] = "#"
		c.Data["Last"] = "#"
	} else {
		c.Data["Next"] = c.URLFor("TagController.List", ":tag_id", tagID, ":page", pageIndex+1)
		c.Data["Last"] = c.URLFor("TagController.List", ":tag_id", tagID, ":page", totalPage)
	}

	c.TplName = "index/list.gohtml"
}
