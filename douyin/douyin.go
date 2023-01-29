package douyin

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

var (
	patternStr       = `http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`
	DefaultUserAgent = `Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1`
	//relRrlStr        = `https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?reflow_source=reflow_page&item_ids=`
	relRrlStr = `https://www.iesdouyin.com/aweme/v1/web/aweme/detail/?aid=1128&version_name=23.5.0&device_platform=android&os_version=2333&aweme_id=`
	//apiStr    = `https://aweme.snssdk.com/aweme/v1/play/?radio=1080p&line=0&video_id=`
	src = rand.NewSource(time.Now().UnixNano())
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
	isDebug bool
	log     *log.Logger
}

func NewDouYin() *DouYin {
	exp, err := regexp.Compile(patternStr)
	if err != nil {
		panic(err)
	}
	return &DouYin{pattern: exp, isDebug: true, log: log.Default()}
}

func (d *DouYin) IsDebug(debug bool) {
	d.isDebug = debug
}

func (d *DouYin) GetRedirectUrl(urlStr string) (string, error) {
	header := http.Header{}
	header.Add("User-Agent", DefaultUserAgent)
	header.Add("Upgrade-Insecure-Requests", "1")

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header = header
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	logs.Info("请求原始内容： %s", string(body))
	exp, err := regexp.Compile("\\d+")
	if err != nil {
		return "", err
	}
	result := exp.FindString(string(body))
	if result == "" {
		return "", errors.New("解析参数失败 ->" + string(body))
	}
	//videoUrl := apiStr + result
	//
	//req, err = http.NewRequest(http.MethodGet, videoUrl, nil)
	//req.Header = header
	//resp, err = http.DefaultTransport.RoundTrip(req)
	//if err != nil {
	//	return "", err
	//}
	//defer resp.Body.Close()
	//
	//body, err = io.ReadAll(resp.Body)
	//if err != nil {
	//	return "", err
	//}
	//log.Println(string(body))

	return relRrlStr + result, nil
}

func (d *DouYin) GetVideoInfo(urlStr string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", DefaultUserAgent)
	cookie := &http.Cookie{Name: "msToken", Value: d.generateRandomStr(107), HttpOnly: true}

	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
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
		if err := recover();err != nil {
			d.printf("解析抖音结果失败 -> [err=%s]", err)
		}
	}()
	urlStr := d.pattern.FindString(shardContent)
	if urlStr == "" {
		return Video{}, errors.New("获取视频链接失败")
	}
	rawUrlStr, err := d.GetRedirectUrl(urlStr)
	if err != nil {
		return Video{}, err
	}
	d.printf("视频链接地址 -> [url=%s]", rawUrlStr)
	body, err := d.GetVideoInfo(rawUrlStr)
	if err != nil {
		return Video{}, err
	}
	d.printf("获取抖音视频成功 -> [resp=%s]", body)
	var result DouYinResult

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		d.printf("解析抖音结果失败 -> [err=%s]", err)
		return Video{}, err
	}
	if len(result.AwemeDetail.Video.PlayAddr.UrlList) == 0 {
		d.printf("解析抖音结果失败 -> [err=%s]", result.FilterDetail.DetailMsg)
		return Video{}, err
	}
	video := Video{
		RawLink:      shardContent,
		VideoRawAddr: urlStr,
		PlayRawAddr:  rawUrlStr,
		Images:       []ImageItem{},
	}

	video.PlayAddr = result.AwemeDetail.Video.PlayAddr.UrlList[0]

	d.printf("视频时长 [duration=%s]", result.AwemeDetail.Duration)
	//获取播放时长，视频有播放时长，图文类无播放时长
	if result.AwemeDetail.Duration > 0 {
		video.VideoType = VideoPlayType
	} else {
		video.VideoType = ImagePlayType
	}
	//获取播放地址
	video.PlayId = result.AwemeDetail.Video.PlayAddr.Uri

	//获取视频唯一id
	d.printf("唯一ID [aweme_id=%s]", result.AwemeDetail.AwemeId)
	video.VideoId = result.AwemeDetail.AwemeId

	//获取封面
	video.Cover = result.AwemeDetail.Video.Cover.UrlList[0]

	//获取原始封面
	video.OriginCover = result.AwemeDetail.Video.OriginCover.UrlList[0]

	video.OriginCoverList = result.AwemeDetail.Video.OriginCover.UrlList
	d.printf("所有原始封面： %+v", video.OriginCoverList)

	//获取音乐地址
	video.MusicAddr = result.AwemeDetail.Music.PlayUrl.UrlList[0]

	//获取作者id
	video.Author.Id = result.AwemeDetail.Author.Uid

	video.Author.ShortId = result.AwemeDetail.Author.ShortId

	video.Author.Nickname = result.AwemeDetail.Author.Nickname

	video.Author.Signature = result.AwemeDetail.Author.Signature

	//获取视频描述
	video.Desc = result.AwemeDetail.Desc

	//回获取作者大头像
	video.Author.AvatarLarger = result.AwemeDetail.Author.AvatarThumb.UrlList[0]

	d.printf("解析后数据 [video=%s]", video.String())
	return video, nil
}

func (d *DouYin) printf(format string, v ...any) {
	if d.isDebug {
		d.log.Printf(format, v...)
	}
}
