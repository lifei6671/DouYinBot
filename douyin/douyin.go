package douyin

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/dop251/goja"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

var (
	patternStr       = `http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`
	DefaultUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36`
	//relRrlStr        = `https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?reflow_source=reflow_page&item_ids=`
	//relRrlStr = `https://www.iesdouyin.com/aweme/v1/web/aweme/detail/?aid=1128&version_name=23.5.0&device_platform=android&os_version=2333&aweme_id=`
	relRrlStr = `https://www.douyin.com/aweme/v1/web/aweme/detail/?aid=1128&version_name=23.5.0&device_platform=android&os_version=2333&aweme_id=`
	//apiStr    = `https://aweme.snssdk.com/aweme/v1/play/?radio=1080p&line=0&video_id=`
	src = rand.NewSource(time.Now().UnixNano())
	//代码来源 https://github.com/wujunwei928/parse-video/blob/main/parser/douyin.go
)

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
	letters      = "ABCDEFGHIGKLMNOPQRSTUVWXYZabcdefghigklmnopqrstuvwxyz0123456789="
)

type DouYin struct {
	pattern *regexp.Regexp
	proxy   string
	isDebug bool
	log     *log.Logger
}

func NewDouYin(proxy string) *DouYin {
	exp, err := regexp.Compile(patternStr)
	if err != nil {
		panic(err)
	}
	return &DouYin{pattern: exp, isDebug: true, log: log.Default(), proxy: proxy}
}

func (d *DouYin) IsDebug(debug bool) {
	d.isDebug = debug
}

func (d *DouYin) generateRandomStr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return string(b)
}

func (d *DouYin) Get(shardContent string) (Video, error) {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("解析抖音结果失败 -> [err=%s]", err)
		}
	}()
	urlStr := d.pattern.FindString(shardContent)
	if urlStr == "" {
		return Video{}, errors.New("获取视频链接失败")
	}

	body, err := d.parseShareUrl(urlStr)
	if err != nil {
		logs.Error("解析抖音结果失败 -> [err=%s]", err)
		return Video{}, err
	}

	logs.Info("获取抖音视频成功 -> [resp=%s]", body)
	var result DouYinResult

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		logs.Error("解析抖音结果失败 -> [err=%s]", err)
		return Video{}, err
	}
	if len(result.VideoData.NwmVideoUrlHQ) == 0 && len(result.VideoData.NwmVideoUrl) == 0 {
		logs.Error("解析抖音结果失败 -> [err=%s]", body)
		return Video{}, errors.New(body)
	}

	video := Video{
		RawLink:      shardContent,
		VideoRawAddr: urlStr,
		PlayRawAddr:  result.Url,
		Images:       []ImageItem{},
	}

	video.PlayAddr = result.VideoData.NwmVideoUrl
	if len(result.VideoData.NwmVideoUrlHQ) > 0 {
		video.PlayAddr = result.VideoData.NwmVideoUrlHQ
	}

	logs.Info("视频时长 [duration=%d]", result.Music.Duration)
	//获取播放时长，视频有播放时长，图文类无播放时长
	if result.Type == "video" {
		video.VideoType = VideoPlayType
	} else {
		video.VideoType = ImagePlayType
	}
	//获取播放地址
	video.PlayId = result.Url

	//获取视频唯一id
	logs.Info("唯一ID [aweme_id=%s]", result.AwemeId)
	video.VideoId = result.AwemeId

	//获取封面
	video.Cover = result.CoverData.Cover.UrlList[0]

	//获取原始封面
	video.OriginCover = result.CoverData.OriginCover.UrlList[0]

	video.OriginCoverList = result.CoverData.OriginCover.UrlList
	logs.Info("所有原始封面： %+v", video.OriginCoverList)

	//获取音乐地址
	video.MusicAddr = result.Music.PlayUrl.UrlList[0]

	//获取作者id
	video.Author.Id = result.Author.Uid

	video.Author.ShortId = result.Author.ShortId

	video.Author.Nickname = result.Author.Nickname

	video.Author.Signature = result.Author.Signature

	//获取视频描述
	video.Desc = result.Desc

	//回获取作者大头像
	video.Author.AvatarLarger = result.Author.AvatarThumb.UrlList[0]

	logs.Info("解析后数据 [video=%s]", video.String())
	return video, nil
}

func (d *DouYin) generateTtwid() string {
	u := "https://ttwid.bytedance.com/ttwid/union/register/"
	data := `{"region":"cn","aid":1768,"needFid":false,"service":"www.ixigua.com","migrate_info":{"ticket":"","source":"node"},"cbUrlProtocol":"https","union":true}`
	resp, err := http.Post(u, "application/json", bytes.NewReader([]byte(data)))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	for _, cookie := range resp.Cookies() {
		return cookie.Value
	}
	return ""
}

func (d *DouYin) GetVideoInfo(reqUrl string) (string, error) {
	client := resty.New()
	res, err := client.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36").
		SetHeader("referer", "https://www.douyin.com/").
		SetHeader("Cookie", fmt.Sprintf(`msToken=%s;odin_tt=324fb4ea4a89c0c05827e18a1ed9cf9bf8a17f7705fcc793fec935b637867e2a5a9b8168c885554d029919117a18ba69; ttwid=1%%7CWBuxH_bhbuTENNtACXoesI5QHV2Dt9-vkMGVHSRRbgY%%7C1677118712%%7C1d87ba1ea2cdf05d80204aea2e1036451dae638e7765b8a4d59d87fa05dd39ff; bd_ticket_guard_client_data=eyJiZC10aWNrZXQtZ3VhcmQtdmVyc2lvbiI6MiwiYmQtdGlja2V0LWd1YXJkLWNsaWVudC1jc3IiOiItLS0tLUJFR0lOIENFUlRJRklDQVRFIFJFUVVFU1QtLS0tLVxyXG5NSUlCRFRDQnRRSUJBREFuTVFzd0NRWURWUVFHRXdKRFRqRVlNQllHQTFVRUF3d1BZbVJmZEdsamEyVjBYMmQxXHJcbllYSmtNRmt3RXdZSEtvWkl6ajBDQVFZSUtvWkl6ajBEQVFjRFFnQUVKUDZzbjNLRlFBNUROSEcyK2F4bXAwNG5cclxud1hBSTZDU1IyZW1sVUE5QTZ4aGQzbVlPUlI4NVRLZ2tXd1FJSmp3Nyszdnc0Z2NNRG5iOTRoS3MvSjFJc3FBc1xyXG5NQ29HQ1NxR1NJYjNEUUVKRGpFZE1Cc3dHUVlEVlIwUkJCSXdFSUlPZDNkM0xtUnZkWGxwYmk1amIyMHdDZ1lJXHJcbktvWkl6ajBFQXdJRFJ3QXdSQUlnVmJkWTI0c0RYS0c0S2h3WlBmOHpxVDRBU0ROamNUb2FFRi9MQnd2QS8xSUNcclxuSURiVmZCUk1PQVB5cWJkcytld1QwSDZqdDg1czZZTVNVZEo5Z2dmOWlmeTBcclxuLS0tLS1FTkQgQ0VSVElGSUNBVEUgUkVRVUVTVC0tLS0tXHJcbiJ9`, d.generateRandomStr(107))).
		Get(reqUrl)
	if err != nil {
		return "", err
	}
	return string(res.Body()), nil
}

func (d *DouYin) GetDetailUrlByVideoId(videoId string) (string, error) {
	postData := &XBogusParam{
		AwemeURL:  fmt.Sprintf("https://www.douyin.com/aweme/v1/web/aweme/detail/?aweme_id=%s&aid=1128&version_name=23.5.0&device_platform=android&os_version=2333", videoId),
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
	}

	xBogus := d.XBogus(postData)

	return postData.AwemeURL + "&X-Bogus=" + xBogus, nil
}

func (d *DouYin) parseShareUrl(shareUrl string) (string, error) {
	proxyURL := d.proxy + "?url=" + shareUrl
	client := resty.New()
	client.SetRedirectPolicy(resty.NoRedirectPolicy())
	res, err := client.R().
		SetHeader("User-Agent", DefaultUserAgent).
		Get(proxyURL)
	// 这里会返回err, auto redirect is disabled
	if err != nil {
		return "", err
	}
	return string(res.Body()), nil
}

type XBogusParam struct {
	AwemeURL  string `json:"aweme_url"`
	UserAgent string `json:"user_agent"`
}

//go:embed X-Bogus.js
var XBogusScript string
var xBogusOnce = sync.Once{}
var vm *goja.Runtime

func (d *DouYin) XBogus(param *XBogusParam) string {
	xBogusOnce.Do(func() {
		vm = goja.New()
		_, err := vm.RunString(XBogusScript)
		if err != nil {
			log.Println("XBogus RunString Err", err)
		}
	})
	sign, ok := goja.AssertFunction(vm.Get("sign"))
	if !ok {
		return ""
	}
	u, err := url.Parse(param.AwemeURL)
	if err != nil {
		log.Println("XBogus RunString AwemeURL Err", err)
		return ""
	}
	res, err := sign(goja.Undefined(), vm.ToValue(u.RawQuery), vm.ToValue(param.UserAgent))
	if err != nil {
		logs.Error("XBogus RunString Sign Err:%s", err)
		return ""
	}
	return res.String()
}
