package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"math"
	"strconv"
)

type IndexController struct {
	web.Controller
}

func (c *IndexController) Index() {

	page := c.Ctx.Input.Param(":page")
	pageIndex := 1
	if page != "" {
		if num, err := strconv.Atoi(page); err == nil {
			pageIndex = int(math.Max(float64(num), float64(pageIndex)))
		}
	}
	list, err := models.NewDouYinVideo().GetList(pageIndex)
	if err != nil {
		logs.Error("获取数据列表失败 -> +%+v", err)
	}
	c.Data["List"] = list
	if pageIndex <= 1 {
		c.Data["Previous"] = "#"
	} else {
		c.Data["Previous"] = c.URLFor("IndexController.Index", ":page", pageIndex-1)
	}
	if list == nil || len(list) == 0 || len(list) < models.PageSize {
		c.Data["Next"] = "#"
	} else {
		c.Data["Next"] = c.URLFor("IndexController.Index", ":page", pageIndex+1)
	}

	c.TplName = "index/index.gohtml"
}
