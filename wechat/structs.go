package wechat

import "encoding/xml"

const (
	WeiXinTextMsgType  WeiXinMsgType = "text"
	WeiXinImageMsgType WeiXinMsgType = "image"
	WeiXinVoiceMsgType WeiXinMsgType = "voice"
	WeiXinVideoMsgType WeiXinMsgType = "video"
	WeiXinMusicMsgType WeiXinMsgType = "music"
	WeiXinNewsMsgType  WeiXinMsgType = "news"
)

type WeiXinMsgType string

type CDATA struct {
	Text string `xml:",cdata"`
}

type PassiveUserReplyMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   CDATA    `xml:"ToUserName"`
	FromUserName CDATA    `xml:"FromUserName"`
	CreateTime   CDATA    `xml:"CreateTime"`
	MsgType      CDATA    `xml:"MsgType"`
	Content      CDATA    `xml:"Content"`
}

func (p *PassiveUserReplyMessage) String() string {
	b, _ := xml.Marshal(p)
	return string(b)
}

type EncryptRequestBody struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string
	Encrypt    string
}
