package douyin

import (
	"errors"
	"github.com/beego/beego/v2/core/logs"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	patternStr = `http[s]?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\(\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+`
	userAgent  = `Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1`
	relRrlStr  = `https://www.iesdouyin.com/web/api/v2/aweme/iteminfo/?item_ids=`
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
	header.Add("User-Agent", userAgent)
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
	return relRrlStr + result, nil
}

func (d *DouYin) GetVideoInfo(urlStr string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
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

func (d *DouYin) Get(shardContent string) (Video, error) {
	urlStr := d.pattern.FindString(shardContent)
	if urlStr == "" {
		return Video{}, errors.New("获取视频链接失败")
	}
	rawUrlStr, err := d.GetRedirectUrl(urlStr)
	if err != nil {
		return Video{}, err
	}
	body, err := d.GetVideoInfo(rawUrlStr)
	if err != nil {
		return Video{}, err
	}
	d.printf("获取抖音视频成功 -> [resp=%s]", body)

	item := gjson.Get(body, "item_list.0")

	res := item.Get("video.play_addr.url_list.0")
	video := Video{
		RawLink: shardContent,
		VideoRawAddr: urlStr,
		PlayRawAddr:  rawUrlStr,
		Images: []ImageItem{},
	}

	if !res.Exists() {
		d.printf("解析抖音视频地址失败 -> [resp=%s]", body)
		return video, errors.New("未找到视频地址 ->" + urlStr)
	}

	video.PlayAddr = strings.ReplaceAll(res.Str, "playwm", "play")
	res = item.Get("duration")
	d.printf("视频时长 [duration=%s]",res.Raw)
	//获取播放时长，视频有播放时长，图文类无播放时长
	if res.Exists() && res.Raw != "0" {
		video.VideoType = VideoPlayType
	} else {
		video.VideoType = ImagePlayType
		res = item.Get("images")
		if res.Exists() && res.IsArray() {
			for _,image := range res.Array() {
				imageRes := image.Get("url_list.0")
				if imageRes.Exists() {
					video.Images = append(video.Images,ImageItem{
						ImageUrl: imageRes.Str,
						ImageId:  image.Get("uri").Str,
					})
				}
			}
		}
	}
	//获取播放地址
	res = item.Get("video.play_addr.uri")
	if res.Exists() {
		video.PlayId = res.Str
	}
	//获取视频唯一id
	res = item.Get("aweme_id")
	d.printf("唯一ID [aweme_id=%s]", res.Raw)
	if res.Exists() {
		video.VideoId = res.Str
	}
	//获取封面
	res = item.Get("video.cover.url_list.0")
	if res.Exists() {
		video.Cover = res.Str
	}
	//获取原始封面
	res = item.Get("video.origin_cover.url_list.0")
	if res.Exists() {
		video.OriginCover = res.Str
	}
	res = item.Get("video.origin_cover.url_list")
	if res.Exists() {
		 res.ForEach(func(key, value gjson.Result) bool {
			 video.OriginCoverList = append(video.OriginCoverList,value.Str)
			 return true
		})
		 d.printf("所有原始封面： %+v", video.OriginCoverList)
	}
	//获取音乐地址
	res = item.Get("music.play_url.url_list.0")
	if res.Exists() {
		video.MusicAddr = res.Str
	}
	//获取作者id
	res = item.Get("author.uid")
	if res.Exists() {
		video.Author.Id = res.Str
	}
	res = item.Get("author.short_id")
	if res.Exists() {
		video.Author.ShortId = res.Str
	}
	res = item.Get("author.nickname")
	if res.Exists() {
		video.Author.Nickname = res.Str
	}
	res = item.Get("author.signature")
	if res.Exists() {
		video.Author.Signature = res.Str
	}
	//获取视频描述
	res = item.Get("desc")
	if res.Exists() {
		video.Desc = res.Str
	}
	//回获取作者大头像
	res = item.Get("author.avatar_larger.url_list.0")
	if res.Exists() {
		video.Author.AvatarLarger = res.Str
	}
	d.printf("解析后数据 [video=%s]",video.String())
	return video, nil
}

func (d *DouYin) printf(format string, v ...interface{}) {
	if d.isDebug {
		d.log.Printf(format,v...)
	}
}
