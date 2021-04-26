package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/service"
	"github.com/lifei6671/douyinbot/wechat"
	"time"
)

var (
	token = web.AppConfig.DefaultString("wechattoken", "")
	key   = web.AppConfig.DefaultString("wechatencodingaeskey", "")
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

	wx := wechat.NewWeiXin(token, key)

	signatureGen := wx.MakeSignature(timestamp, nonce)

	if signatureGen == signature {
		_ = c.Ctx.Output.Body([]byte(echostr))
	} else {
		_ = c.Ctx.Output.Body([]byte("false"))
	}
	c.StopRun()
}

func (c *WeiXinController) Dispatch() {
	doc, err := xmlquery.Parse(bytes.NewReader(c.Ctx.Input.RequestBody))
	if err != nil {
		logs.Error("解析微信消息失败 -> %+v", err)
		_ = c.Ctx.Output.Body([]byte("success"))
		c.StopRun()
	}
	node, err := xmlquery.Query(doc, "//MsgType")
	if err != nil {
		logs.Error("解析微信消息失败 -> %+v", err)
		_ = c.Ctx.Output.Body([]byte("success"))
		c.StopRun()
		return
	}
	var toUserName, fromUserName string

	toUserNameNode, err := xmlquery.Query(doc, "//ToUserName")
	if err != nil || toUserNameNode == nil {
		logs.Error("解析微信消息失败 -> %s - %+v", string(c.Ctx.Input.RequestBody), err)
		_ = c.Ctx.Output.Body([]byte("success"))
		c.StopRun()
		return
	}
	toUserName = toUserNameNode.InnerText()

	fromUserNameNode, err := xmlquery.Query(doc, "//FromUserName")
	if err != nil || fromUserNameNode == nil {
		logs.Error("解析微信消息失败 -> %s - %+v", string(c.Ctx.Input.RequestBody), err)
		_ = c.Ctx.Output.Body([]byte("success"))
		c.StopRun()
		return
	}
	fromUserName = fromUserNameNode.InnerText()

	if node.InnerText() == string(wechat.WeiXinTextMsgType) {
		node, err = xmlquery.Query(doc, "//Content")
		if err != nil || node == nil {
			c.Data["xml"] = wechat.PassiveUserReplyMessage{
				ToUserName:   wechat.Value(fromUserName),
				FromUserName: wechat.Value(toUserName),
				CreateTime:   wechat.Value(fmt.Sprintf("%d", time.Now().Unix())),
				MsgType:      wechat.Value(string(wechat.WeiXinTextMsgType)),
				Content:      wechat.Value("解析消息内容失败"),
			}
			_ = c.ServeXML()
			return
		}
		content := node.InnerText()
		service.Push(context.Background(), content)

		resp := wechat.PassiveUserReplyMessage{
			ToUserName:   wechat.Value(fromUserName),
			FromUserName: wechat.Value(toUserName),
			CreateTime:   wechat.Value(fmt.Sprintf("%d", time.Now().Unix())),
			MsgType:      wechat.Value(string(wechat.WeiXinTextMsgType)),
			Content:      wechat.Value("处理成功"),
		}
		c.Data["xml"] = resp
		_ = c.ServeXML()
		return
	}
	resp := wechat.PassiveUserReplyMessage{
		ToUserName:   wechat.Value(fromUserName),
		FromUserName: wechat.Value(toUserName),
		CreateTime:   wechat.Value(fmt.Sprintf("%d", time.Now().Unix())),
		MsgType:      wechat.Value(string(wechat.WeiXinTextMsgType)),
		Content:      wechat.Value("不支持的消息类型"),
	}
	c.Data["xml"] = resp

	_ = c.ServeXML()
	return

}
