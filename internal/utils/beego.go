package utils

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web/context"
)

// CacheHeader 设置缓存响应头
func CacheHeader(ctx *context.BeegoOutput, t time.Time, minAge, maxAge int) {
	lastTime := time.Date(t.Year(), t.Month(), t.Day(), 1, 0, 0, 0, t.Location()).UTC().Format(http.TimeFormat)
	ctx.Header("Cache-Control", fmt.Sprintf("max-age=%d, s-maxage=%d", minAge, maxAge))
	ctx.Header("Cloudflare-CDN-Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
	ctx.Header("Date", lastTime)
	ctx.Header("Last-Modified", lastTime)
}

// IfLastModified 判断是否和当前时间匹配
func IfLastModified(ctx *context.BeegoInput, t time.Time) error {
	lastTime := time.Date(t.Year(), t.Month(), t.Day(), 1, 0, 0, 0, t.Location()).UTC().Format(http.TimeFormat)
	modified := ctx.Header("If-Modified-Since")
	if modified != "" && lastTime == modified {
		return nil
	}
	return errors.New("Last-Modified not supported")
}
