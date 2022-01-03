package utils

import (
	"errors"
	"net/url"
	"strconv"
)
// ParseExpireUnix 从URL中解析过期时间
func ParseExpireUnix(s string) (int,error) {
	uri, err := url.ParseRequestURI(s)
	if err != nil {
		return 0,err
	}
	if v := uri.Query().Get("x-expires"); v != "" {
		expire, err := strconv.Atoi(v)
		if err != nil {
			return 0,err
		}
		return expire,nil
	}
	return 0, errors.New("url is empty")
}
