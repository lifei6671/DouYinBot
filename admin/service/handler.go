package service

import (
	"fmt"
	"github.com/lifei6671/douyinbot/wechat"
	"time"
)

var (
	handlers = make(map[string]WechatHandler)
)

type WechatHandler func(*wechat.TextRequestBody) (wechat.PassiveUserReplyMessage, error)

func UserRegisterHandler(textRequestBody *wechat.TextRequestBody) (wechat.PassiveUserReplyMessage, error) {

	return wechat.PassiveUserReplyMessage{
		ToUserName:   wechat.Value(textRequestBody.FromUserName),
		FromUserName: wechat.Value(textRequestBody.ToUserName),
		CreateTime:   wechat.Value(fmt.Sprintf("%d", time.Now().Unix())),
		MsgType:      wechat.Value(string(wechat.WeiXinTextMsgType)),
		Content:      wechat.Value("注册格式: 注册#账号#密码#邮箱地址"),
	}, nil
}

func RegisterHandler(keyword string, handler WechatHandler) {
	handlers[keyword] = handler
}

func GetHandler(keyword string) WechatHandler {
	if handler, ok := handlers[keyword]; ok {
		return handler
	}
	return nil
}

func init() {
	handlers["注册"] = UserRegisterHandler
}
