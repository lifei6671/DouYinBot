package douyin

import (
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)
type VideoType int

const (
	//VideoPlayType 视频类
	VideoPlayType VideoType = 0
	//ImagePlayType 图文类
	ImagePlayType VideoType = 1
)
type Video struct {
	VideoId 	 string
	PlayId       string
	PlayAddr     string
	VideoRawAddr string
	PlayRawAddr  string
	Cover        string
	OriginCover  string
	MusicAddr    string
	Desc         string
	Author       struct {
		Id           string
		ShortId      string
		Nickname     string
		AvatarLarger string
		Signature    string
	}
	Images []ImageItem
	VideoType VideoType
}

type ImageItem struct {
	ImageUrl string
	ImageId string
}

func (v *Video) GetFilename() string {
	return v.PlayId + ".mp4"
}

//Download 下载文件到指定目录
func (v *Video) Download(filename string) (string, error) {
	name := filepath.Join(v.Author.Id, v.GetFilename())
	f, err := os.Stat(filename)
	if err == nil && f.IsDir() {
		filename = filepath.Join(filename, name)
	}
	dir := filepath.Dir(filename)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0655); err != nil {
			return "", err
		}
	}
	//如果是图片类，则将图片下载到指定目录
	if v.VideoType == ImagePlayType {
		for _,image := range v.Images {
			imagePath := filepath.Join(dir,v.VideoId,image.ImageId)
			req,err := http.NewRequest(http.MethodGet, image.ImageUrl,nil)
			if err != nil {
				logs.Error("下载图像出错 -> [play_id=%s] [image_url=%s]", v.PlayId,image.ImageUrl)
				continue
			}
			b,err := io.ReadAll(req.Body)
			if err != nil {
				logs.Error("解析图像出错 -> [play_id=%s] [image_url=%s]", v.PlayId,image.ImageUrl)
				continue
			}
			_ = req.Body.Close()
			err = ioutil.WriteFile(imagePath,b, 0655)
			if err != nil {
				logs.Error("保存图像出错 -> [play_id=%s] [image_url=%s]", v.PlayId,image.ImageUrl)
				continue
			}
			time.Sleep(time.Microsecond * 110)
		}
		//如果是图文，需要将音频和图像放入一个目录
		filename = filepath.Join(dir, v.VideoId, name)
	}
	req, err := http.NewRequest(http.MethodGet, v.PlayAddr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()


	f1, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f1.Close()
	_, err = io.Copy(f1, resp.Body)
	return filename, err
}

//GetDownloadUrl 获取下载链接
func (v *Video) GetDownloadUrl() (string, error) {
	req, err := http.NewRequest(http.MethodGet, v.PlayAddr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	lv := resp.Header.Get("Location")

	return lv, nil
}

func (v *Video) String() string {
	b, err := json.Marshal(v)
	if err != nil {
		logs.Error("编码失败 -> %s", err)
	} else {
		return string(b)
	}
	return fmt.Sprintf("%+v", *v)
}
