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
	token = web.AppConfig.DefaultString("wechattoken", "")
	key   = web.AppConfig.DefaultString("wechatencodingaeskey", "")
	appid = web.AppConfig.DefaultString("wechatappid", "")
)

type WeiXinController struct {
	web.Controller
}

// Index 严重是否是微信请求
func (c *WeiXinController) Index() {
	timestamp := c.Ctx.Input.Query("timestamp")
	signature := c.Ctx.Input.Query("signature")
	nonce := c.Ctx.Input.Query("nonce")
	echostr := c.Ctx.Input.Query("echostr")

	wx := wechat.NewWeiXin(appid, token, key)

	signatureGen := wx.MakeSignature(timestamp, nonce)

	if signatureGen == signature {
		_ = c.Ctx.Output.Body([]byte(echostr))
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

	wx := wechat.NewWeiXin(appid, token, key)

	textRequestBody := &wechat.TextRequestBody{}

	logs.Info("微信请求 ->", string(c.Ctx.Input.RequestBody))
	if encryptType == wechat.EncryptTypeAES {
		requestBody := &wechat.EncryptRequestBody{}
		if err := xml.Unmarshal(c.Ctx.Input.RequestBody, requestBody); err != nil {
			logs.Error("解析微信消息失败 -> %+v", err)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}
		if !wx.ValidateMsg(timestamp, nonce, requestBody.Encrypt, msgSignature) {
			logs.Error("解析微信消息失败 -> %+v", msgSignature)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}
		var err error

		textRequestBody, err = wx.ParseEncryptRequestBody(timestamp, nonce, msgSignature, c.Ctx.Input.RequestBody)
		if err != nil {
			logs.Error("解析微信消息失败 -> %+v", err)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}
	} else {

		err := xml.Unmarshal(c.Ctx.Input.RequestBody, textRequestBody)
		if err != nil {
			logs.Error("解析微信消息失败 -> %+v", msgSignature)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return
		}

	}

	if textRequestBody.MsgType == string(wechat.WeiXinTextMsgType) {
		if textRequestBody.Content == "" {
			c.response(wx, textRequestBody, "解析消息失败")
			return
		}
		service.Push(context.Background(), textRequestBody.Content)

		c.response(wx, textRequestBody, "处理成功")
		return
	}
	c.response(wx, textRequestBody, "不支持的消息类型")
}

func (c *WeiXinController) response(wx *wechat.WeiXin, textRequestBody *wechat.TextRequestBody, content string) error {
	nonce := c.Ctx.Input.Query("nonce")
	timestamp := c.Ctx.Input.Query("timestamp")
	encryptType := c.Ctx.Input.Query("encrypt_type")

	if encryptType == wechat.EncryptTypeAES {
		c.Data["xml"] = wechat.PassiveUserReplyMessage{
			ToUserName:   wechat.Value(textRequestBody.FromUserName),
			FromUserName: wechat.Value(textRequestBody.ToUserName),
			CreateTime:   wechat.Value(fmt.Sprintf("%d", time.Now().Unix())),
			MsgType:      wechat.Value(string(wechat.WeiXinTextMsgType)),
			Content:      wechat.Value(content),
		}
		return c.ServeXML()
	} else {
		body, err := wx.MakeEncryptResponseBody(textRequestBody.ToUserName, textRequestBody.FromUserName, content, nonce, timestamp)
		if err != nil {
			logs.Error("解析微信消息失败 -> %+v", textRequestBody)
			_ = c.Ctx.Output.Body([]byte("success"))
			c.StopRun()
			return err
		}
		err = c.Ctx.Output.Body(body)
		c.StopRun()
		return err
	}
}
