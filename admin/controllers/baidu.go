package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/baidu"
	"net/http"
	"time"
)

type BaiduController struct {
	web.Controller
	bd *baidu.Netdisk
}

func (c *BaiduController) Prepare() {
	if !web.AppConfig.DefaultBool("baidunetdiskenable", false) {
		c.Ctx.Output.Body([]byte("网站未开启百度网盘接入功能！"))
		c.StopRun()
		return
	}
	appId := web.AppConfig.DefaultString("baiduappid", "")
	appKey := web.AppConfig.DefaultString("baiduappkey", "")
	secretKey := web.AppConfig.DefaultString("baidusecretkey", "")
	signKey := web.AppConfig.DefaultString("baidusignkey", "")

	bd := baidu.NewNetdisk(appId, appKey, secretKey, signKey)
	c.bd = bd
}
func (c *BaiduController) Index() {
	registeredUrl := web.AppConfig.DefaultString("baiduregisteredurl", "")
	authorizeUrl := c.bd.AuthorizeURI(registeredUrl)
	c.Redirect(authorizeUrl, http.StatusFound)
	c.StopRun()
}

func (c *BaiduController) Authorize() {
	code := c.Ctx.Input.Query("code")
	if code == "" {
		c.Ctx.Output.Body([]byte("获取百度网盘授权信息失败！"))
		c.StopRun()
		return
	}
	registeredUrl := web.AppConfig.DefaultString("baiduregisteredurl", "")

	token, err := c.bd.GetAccessToken(code, registeredUrl)
	if err != nil {
		logs.Error("百度网盘授权失败 -> [code=%s] error=%+v", code, err)
		c.Ctx.Output.Body([]byte("获取百度网盘授权信息失败！"))
		c.StopRun()
		return
	}
	userInfo, err := c.bd.UserInfo()
	if err != nil {
		logs.Error("百度网盘用户信息失败 -> [code=%s] error=%+v", code, err)
		c.Ctx.Output.Body([]byte("获取百度网盘用户信息失败！"))
		c.StopRun()
		return
	}
	user := models.NewBaiduToken()
	user.BaiduId = userInfo.UserId
	user.BaiduName = userInfo.BaiduName
	user.VipType = userInfo.VipType
	user.NetdiskName = userInfo.NetdiskName
	user.AvatarUrl = userInfo.AvatarUrl
	user.AccessToken = token.AccessToken
	user.RefreshToken = token.RefreshToken
	user.ExpiresIn = token.ExpiresIn
	user.Scope = token.Scope
	user.Created = time.Unix(token.CreateAt, 0)
	user.RefreshTokenCreateAt = time.Unix(token.RefreshTokenCreateAt, 0)
	err = user.Save()
	if err != nil {
		logs.Error("百度网盘用户信息失败 -> [code=%s] error=%+v", code, err)
		c.Ctx.Output.Body([]byte("保存百度网盘用户信息失败！"))
		c.StopRun()
		return
	}
	logs.Info("百度网盘授权成功 -> [code=%s] user:%+v", code, user)
	c.Ctx.Output.Body([]byte("百度网盘授权成功！"))
	c.StopRun()
	return
}
