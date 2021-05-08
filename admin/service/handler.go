package service

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/wechat"
	"net/url"
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

func BindBaiduNetdiskHandler(textRequestBody *wechat.TextRequestBody) (wechat.PassiveUserReplyMessage, error) {
	message := wechat.PassiveUserReplyMessage{
		ToUserName:   wechat.Value(textRequestBody.FromUserName),
		FromUserName: wechat.Value(textRequestBody.ToUserName),
		CreateTime:   wechat.Value(fmt.Sprintf("%d", time.Now().Unix())),
		MsgType:      wechat.Value(string(wechat.WeiXinTextMsgType)),
	}

	if !web.AppConfig.DefaultBool("baidunetdiskenable", false) {
		message.Content = wechat.Value("网站未开启百度网盘接入功能")
		return message, nil
	}
	if !models.NewUser().ExistByWechatId(textRequestBody.FromUserName) {
		message.Content = wechat.Value("你不是已注册用户不能绑定百度网盘")
		return message, nil
	}
	registeredUrl := web.AppConfig.DefaultString("baiduregisteredurl", "")
	uri, err := url.Parse(registeredUrl)
	if err != nil {
		logs.Error("解析注册回调地址失败 -> %s - %+v", registeredUrl, err)
		message.Content = wechat.Value("生成百度网盘绑定数据失败")
		return message, nil
	}
	message.Content = wechat.Value(uri.Scheme + "://" + uri.Host + "/baidu/authorize?wid=" + textRequestBody.FromUserName)
	return message, nil
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
	handlers["1"] = UserRegisterHandler
	handlers["2"] = BindBaiduNetdiskHandler

}
