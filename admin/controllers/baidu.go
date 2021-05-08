package controllers

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/admin/service"
	"github.com/lifei6671/douyinbot/baidu"
	"net/http"
	"time"
)

var (
	bd *baidu.Netdisk
)

type BaiduController struct {
	web.Controller
	display string
}

func (c *BaiduController) Prepare() {
	if !web.AppConfig.DefaultBool("baidunetdiskenable", false) {
		_ = c.Ctx.Output.Body([]byte("网站未开启百度网盘接入功能！"))
		c.StopRun()
		return
	}
	c.display = "mobile"
	if !service.IsMobile(c.Ctx.Input.UserAgent()) {
		c.display = "page"
	}
}
func (c *BaiduController) Index() {
	wid := c.Ctx.Input.Query("wid")
	if wid == "" {
		_ = c.Ctx.Output.Body([]byte("参数错误"))
		c.StopRun()
		return
	}
	if err := c.SetSession("wid", wid); err != nil {
		_ = c.Ctx.Output.Body([]byte("保存参数失败"))
		c.StopRun()
		return
	}

	registeredUrl := web.AppConfig.DefaultString("baiduregisteredurl", "")

	authorizeUrl := bd.AuthorizeURI(registeredUrl, c.display)
	c.Redirect(authorizeUrl, http.StatusFound)
	c.StopRun()
}

func (c *BaiduController) Authorize() {
	code := c.Ctx.Input.Query("code")
	if code == "" {
		_ = c.Ctx.Output.Body([]byte("获取百度网盘授权信息失败！"))
		c.StopRun()
		return
	}
	wid, ok := c.GetSession("wid").(string)
	if !ok {
		_ = c.Ctx.Output.Body([]byte("授权失败请重新发起授权"))
		return
	}

	registeredUrl := web.AppConfig.DefaultString("baiduregisteredurl", "")

	token, err := bd.GetAccessToken(code, registeredUrl)
	if err != nil {
		logs.Error("百度网盘授权失败 -> [code=%s] error=%s", code, err)
		_ = c.Ctx.Output.Body([]byte("获取百度网盘授权信息失败！"))
		c.StopRun()
		return
	}
	userInfo, err := bd.UserInfo()
	if err != nil {
		logs.Error("百度网盘用户信息失败 -> [code=%s] error=%+v", code, err)
		_ = c.Ctx.Output.Body([]byte("获取百度网盘用户信息失败！"))
		c.StopRun()
		return
	}

	user, err := models.NewUser().FirstByWechatId(wid)
	if err != nil {
		_ = c.Ctx.Output.Body([]byte("您不是已注册用户不能绑定百度网盘"))
		return
	}
	user.BaiduId = userInfo.UserId
	if err := user.Update("baidu_id"); err != nil {
		logs.Error("更新用户BaiduId失败 -> %+v", err)
		_ = c.Ctx.Output.Body([]byte("绑定用户网盘失败，请重试"))
		return
	}

	baiduUser := models.NewBaiduToken()
	baiduUser.BaiduId = userInfo.UserId
	baiduUser.BaiduName = userInfo.BaiduName
	baiduUser.VipType = userInfo.VipType
	baiduUser.NetdiskName = userInfo.NetdiskName
	baiduUser.AvatarUrl = userInfo.AvatarUrl
	baiduUser.AccessToken = token.AccessToken
	baiduUser.RefreshToken = token.RefreshToken
	baiduUser.ExpiresIn = token.ExpiresIn
	baiduUser.Scope = token.Scope
	baiduUser.Created = time.Unix(token.CreateAt, 0)
	baiduUser.RefreshTokenCreateAt = time.Unix(token.RefreshTokenCreateAt, 0)
	err = baiduUser.Save()
	if err != nil {
		logs.Error("百度网盘用户信息失败 -> [code=%s] error=%+v", code, err)
		_ = c.Ctx.Output.Body([]byte("保存百度网盘用户信息失败！"))
		c.StopRun()
		return
	}
	logs.Info("百度网盘授权成功 -> [code=%s] baiduUser:%+v", code, baiduUser)
	_ = c.Ctx.Output.Body([]byte("百度网盘授权成功！"))
	c.StopRun()
	return
}

func init() {
	if !web.AppConfig.DefaultBool("baidunetdiskenable", false) {
		return
	}
	appId := web.AppConfig.DefaultString("baiduappid", "")
	appKey := web.AppConfig.DefaultString("baiduappkey", "")
	secretKey := web.AppConfig.DefaultString("baidusecretkey", "")
	signKey := web.AppConfig.DefaultString("baidusignkey", "")

	bd = baidu.NewNetdisk(appId, appKey, secretKey, signKey)
	bd.IsDebug(true)
}
