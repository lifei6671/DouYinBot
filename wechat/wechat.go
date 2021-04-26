package wechat

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strings"
)

type WeiXin struct {
	token       string
	encodingKey string
	aesKey      []byte
}

func NewWeiXin(token, key string) *WeiXin {
	return &WeiXin{
		token:       token,
		encodingKey: key,
	}
}

func (w *WeiXin) MakeSignature(timestamp, nonce string) string { //本地计算signature
	si := []string{w.token, timestamp, nonce}
	sort.Strings(si)            //字典序排序
	str := strings.Join(si, "") //组合字符串
	s := sha1.New()             //返回一个新的使用SHA1校验的hash.Hash接口
	_, err := io.WriteString(s, str)
	if err != nil {
		return ""
	}
	//WriteString函数将字符串数组str中的内容写入到s中
	return fmt.Sprintf("%x", s.Sum(nil))
}

func (w *WeiXin) EncodingAESKey2AESKey() []byte {
	if w.aesKey == nil || len(w.aesKey) == 0 {
		data, _ := base64.StdEncoding.DecodeString(w.encodingKey + "=")
		w.aesKey = data
	}
	b := make([]byte, len(w.aesKey))
	copy(b, w.aesKey)
	return b
}
