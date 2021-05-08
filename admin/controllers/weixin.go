package controllers

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/service"
	"github.com/lifei6671/douyinbot/wechat"
	"time"
)

var (
	token            = web.AppConfig.DefaultString("wechattoken", "")
	key              = web.AppConfig.DefaultString("wechatencodingaeskey", "")
	appId            = web.AppConfig.DefaultString("wechatappid", "")
	autoReplyContent = "回复关键字：\n1：注册账号\n2：绑定百度网盘"
)

type WeiXinController struct {
	web.Controller
	wx     *wechat.WeiXin
	body   *wechat.TextRequestBody
	domain string
}

func (c *WeiXinController) Prepare() {
	c.wx = wechat.NewWeiXin(appId, token, key)
	c.domain = c.Ctx.Input.Scheme() + "://" + c.Ctx.Input.Host()
}

// Index 验证是否是微信请求
func (c *WeiXinController) Index() {
	timestamp := c.Ctx.Input.Query("timestamp")
	signature := c.Ctx.Input.Query("signature")
	nonce := c.Ctx.Input.Query("nonce")
	echoStr := c.Ctx.Input.Query("echoStr")

	signatureGen := c.wx.MakeSignature(timestamp, nonce)

	if signatureGen == signature {
		_ = c.Ctx.Output.Body([]byte(echoStr))
	} else {
		_ = c.Ctx.Output.Body([]byte("false"))
	}
	c.StopRun()
}

func (c *WeiXinController) Dispatch() {
	encryptType := c.Ctx.Input.Query("encrypt_type")
	msgSignature := c.Ctx.Input.Query("msg_signature")
	nonce := c.Ctx.Input.Query("nonce")
	timestamp := c.Ctx.Input.Query("timestamp")

	logs.Info("微信请求 ->", string(c.Ctx.Input.RequestBody))
	if encryptType == wechat.EncryptTypeAES {
		requestBody := &wechat.EncryptRequestBody{}
		if err := xml.Unmarshal(c.Ctx.Input.RequestBody, requestBody); err != nil {
			logs.Error("解析微信消息失败 -> %+v", err)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}
		if !c.wx.ValidateMsg(timestamp, nonce, requestBody.Encrypt, msgSignature) {
			logs.Error("解析微信消息失败 -> %+v", msgSignature)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}
		var err error

		c.body, err = c.wx.ParseEncryptRequestBody(timestamp, nonce, msgSignature, c.Ctx.Input.RequestBody)
		if err != nil {
			logs.Error("解析微信消息失败 -> %+v", err)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}
	} else {
		var textRequestBody wechat.TextRequestBody
		err := xml.Unmarshal(c.Ctx.Input.RequestBody, &textRequestBody)
		if err != nil {
			logs.Error("解析微信消息失败 -> %+v", msgSignature)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}
		c.body = &textRequestBody
	}

	if c.body.MsgType == string(wechat.WeiXinTextMsgType) {
		if c.body.Content == "" {
			_ = c.response("解析消息失败")
			return
		}
		if handler := service.GetHandler(c.body.Content); handler != nil {
			if resp, err := handler(c.body); err != nil {
				_ = c.response("处理失败")
			} else {
				c.responseBody(resp)
			}
			return
		}
		if err := service.Register(c.body.Content, c.body.FromUserName); err != service.ErrNoUserRegister {
			if err != nil {
				_ = c.response(err.Error())
			} else {
				_ = c.response("注册成功")
			}
			return
		}
		service.Push(context.Background(), service.MediaContent{
			Content: c.body.Content,
			UserId:  c.body.FromUserName,
		})

		_ = c.response("处理成功")
		return
	} else if c.body.MsgType == string(wechat.WeiXinEventMsgType) {
		//如果是推送的订阅事件
		if c.body.Event == wechat.WeiXinSubscribeEvent {
			_ = c.response(autoReplyContent)
		}

	}
	_ = c.response("不支持的消息类型")
}

func (c *WeiXinController) responseBody(resp wechat.PassiveUserReplyMessage) {
	nonce := c.Ctx.Input.Query("nonce")
	timestamp := c.Ctx.Input.Query("timestamp")
	encryptType := c.Ctx.Input.Query("encrypt_type")

	if encryptType == wechat.EncryptTypeAES {
		c.Data["xml"] = resp
		_ = c.ServeXML()
	} else {
		body, err := c.wx.MakeEncryptResponseBody(resp.FromUserName.Text, resp.ToUserName.Text, resp.Content.Text, nonce, timestamp)
		if err != nil {
			logs.Error("解析微信消息失败 -> %+v", resp)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
		}
		err = c.Ctx.Output.Body(body)
	}
	c.StopRun()
}

func (c *WeiXinController) response(content string) error {
	nonce := c.Ctx.Input.Query("nonce")
	timestamp := c.Ctx.Input.Query("timestamp")
	encryptType := c.Ctx.Input.Query("encrypt_type")

	if encryptType == wechat.EncryptTypeAES {
		c.Data["xml"] = wechat.PassiveUserReplyMessage{
			ToUserName:   wechat.Value(c.body.FromUserName),
			FromUserName: wechat.Value(c.body.ToUserName),
			CreateTime:   wechat.Value(fmt.Sprintf("%d", time.Now().Unix())),
			MsgType:      wechat.Value(string(wechat.WeiXinTextMsgType)),
			Content:      wechat.Value(content),
		}
		return c.ServeXML()
	} else {
		body, err := c.wx.MakeEncryptResponseBody(c.body.ToUserName, c.body.FromUserName, content, nonce, timestamp)
		if err != nil {
			logs.Error("解析微信消息失败 -> %+v", c.body)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return err
		}
		err = c.Ctx.Output.Body(body)
		c.StopRun()
		return err
	}
}
