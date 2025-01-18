package douyin

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"regexp"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/go-resty/resty/v2"

	"github.com/lifei6671/douyinbot/internal/utils"
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
	//抖音抓取代理
	proxy string
	//代理账号
	username string
	//代理密码
	password string
	isDebug  bool
	log      *log.Logger
}

func NewDouYin(proxy, username, password string) *DouYin {
	exp, err := regexp.Compile(patternStr)
	if err != nil {
		panic(err)
	}
	return &DouYin{
		pattern: exp, isDebug: true, log: log.Default(), proxy: proxy,
		username: username,
		password: password,
	}
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

	//解析图片
	if len(result.Images) > 0 {
		for _, image := range result.Images {
			video.Images = append(video.Images, ImageItem{
				ImageUrl: utils.First(image.URLList),
				ImageId:  image.URI,
			})
		}
	}

	//获取封面
	video.Cover = utils.First(result.CoverData.Cover.UrlList)

	//获取原始封面
	video.OriginCover = utils.First(result.CoverData.Cover.UrlList)

	video.OriginCoverList = result.CoverData.Cover.UrlList
	logs.Info("所有原始封面： %+v", video.OriginCoverList)

	//获取音乐地址
	video.MusicAddr = utils.First(result.Music.PlayUrl.UrlList)

	//获取作者id
	video.Author.Id = result.Author.Uid

	video.Author.ShortId = result.Author.ShortId

	video.Author.Nickname = result.Author.Nickname

	video.Author.Signature = result.Author.Signature

	//获取视频描述
	video.Desc = result.Desc

	//回获取作者大头像
	video.Author.AvatarLarger = utils.First(result.Author.AvatarThumb.UrlList)

	logs.Info("解析后数据 [video=%s]", video.String())
	return video, nil
}

func (d *DouYin) GetVideoInfo(reqUrl string) (string, error) {
	return d.parseShareUrl(reqUrl)
}

func (d *DouYin) parseShareUrl(shareUrl string) (string, error) {
	proxyURL := d.proxy + "?url=" + shareUrl
	client := resty.New()

	log.Println(d.username, d.password)
	res, err := client.R().
		SetHeader("User-Agent", DefaultUserAgent).
		SetBasicAuth(d.username, d.password).
		Get(proxyURL)

	// 这里会返回err, auto redirect is disabled
	if err != nil {
		return "", err
	}
	return string(res.Body()), nil
}
