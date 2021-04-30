package controllers

import "github.com/beego/beego/v2/server/web"

type LoginController struct {
	web.Controller
}

func (c *LoginController) Login() {
	if c.Ctx.Input.IsGet() {

	} else {

	}
}
