package douyin

import (
	"errors"
	"github.com/tidwall/gjson"
	"io"
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
}

func NewDouYin() *DouYin {
	exp, err := regexp.Compile(patternStr)
	if err != nil {
		panic(err)
	}
	return &DouYin{pattern: exp}
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
	item := gjson.Get(body, "item_list.0")

	res := item.Get("video.play_addr.url_list.0")
	video := Video{
		VideoRawAddr: urlStr,
		PlayRawAddr:  rawUrlStr,
	}

	if !res.Exists() {
		return video, errors.New("未找到视频地址 ->" + urlStr)
	}

	video.PlayAddr = strings.ReplaceAll(res.Str, "playwm", "play")
	res = item.Get("video.play_addr.uri")
	if res.Exists() {
		video.PlayId = res.Str
	}
	res = item.Get("video.cover.url_list.0")
	if res.Exists() {
		video.Cover = res.Str
	}
	res = item.Get("video.origin_cover.url_list.0")
	if res.Exists() {
		video.OriginCover = res.Str
	}
	res = item.Get("music.play_url.url_list.0")
	if res.Exists() {
		video.MusicAddr = res.Str
	}
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

	res = item.Get("desc")
	if res.Exists() {
		video.Desc = res.Str
	}
	res = item.Get("author.avatar_larger.url_list.0")
	if res.Exists() {
		video.Author.AvatarLarger = res.Str
	}
	return video, nil
}
