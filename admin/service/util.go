package service

import (
	"github.com/beego/beego/v2/core/logs"
	"regexp"
)

func IsMobile(userAgent string) bool {
	reg := "(iphone|MicroMessenger|ios|android|mini|mobile|mobi|Nokia|Symbian|iPod|iPad|Windows\\s+Phone|MQQBrowser|wp7|wp8|UCBrowser7|UCWEB|360\\s+Aphone\\s+Browser)"
	isMobile, err := regexp.Match(reg, []byte(userAgent))
	if err != nil {
		logs.Error("匹配User-Agent失败 -> %+v", err)
	}
	return isMobile
}
