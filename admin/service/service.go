package service

import (
	"context"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/douyin"
	"github.com/lifei6671/douyinbot/qiniu"
	"os"
	"strings"
)

var (
	videoShareChan = make(chan string, 100)
	accessKey      = ""
	secretKey      = ""
	bucketName     = ""
	domain         = ""
	savepath       = ""
)

func Push(ctx context.Context, content string) {
	select {
	case videoShareChan <- content:
	case <-ctx.Done():
	}
}

func Run(ctx context.Context) (err error) {
	accessKey, err = web.AppConfig.String("qiuniuaccesskey")
	if err != nil {
		logs.Error("获取七牛配置失败 -> [qiuniuaccesskey] - %+v", err)
		return err
	}
	secretKey, err = web.AppConfig.String("qiuniusecretkey")
	if err != nil {
		logs.Error("获取七牛配置失败 -> [qiuniusecretkey] - %+v", err)
		return err
	}
	bucketName, err = web.AppConfig.String("qiuniubucketname")
	if err != nil {
		logs.Error("获取七牛配置失败 -> [qiuniubucketname] - %+v", err)
		return err
	}
	domain, err = web.AppConfig.String("qiniudoamin")
	if err != nil {
		logs.Error("获取七牛配置失败 -> [qiniudoamin] - %+v", err)
		return err
	}
	savepath, err = web.AppConfig.String("auto-save-path")
	if err != nil {
		logs.Error("获取本地储存目录失败 ->[auto-save-path] %+v", err)
		return err
	}
	go execute(ctx)
	go execute(ctx)
	go execute(ctx)
	go execute(ctx)
	return nil
}

func execute(ctx context.Context) {
	dy := douyin.NewDouYin()
	bucket := qiniu.NewBucket(accessKey, secretKey)
	for {
		select {
		case content, ok := <-videoShareChan:
			if !ok {
				return
			}
			logs.Info("开始解析抖音视频任务 -> %s", content)
			video, err := dy.Get(content)
			if err != nil {
				logs.Error("解析抖音视频地址失败 -> 【%s】- %+v", content, err)
				continue
			}
			p, err := video.Download(savepath)
			if err != nil {
				logs.Error("下载抖音视频失败 -> 【%s】- %+v", content, err)
				continue
			}
			name := strings.TrimPrefix(p, savepath)

			err = bucket.UploadFile(bucketName, name, p)
			if err != nil {
				logs.Error("上传文件到七牛储存空间失败 -> 【%s】 - %+v", content, err)
				_ = os.Remove(p)
				continue
			}

			m := models.DouYinVideo{
				Nickname:         video.Author.Nickname,
				Signature:        video.Author.Signature,
				AvatarLarger:     video.Author.AvatarLarger,
				AuthorId:         video.Author.Id,
				AuthorShortId:    video.Author.ShortId,
				VideoRawPlayAddr: video.VideoRawAddr,
				VideoPlayAddr:    video.PlayAddr,
				VideoId:          video.PlayId,
				VideoCover:       video.OriginCover,
				VideoLocalAddr:   "/video/" + name,
				VideoBackAddr:    domain + name,
				Desc:             video.Desc,
			}
			if err := m.Save(); err != nil {
				logs.Error("保存视频到数据库失败 -> 【%s】 - %+v", content, err)
				continue
			}
			logs.Info("解析抖音视频成功 -> 【%s】- %s", content, m.VideoBackAddr)

		case <-ctx.Done():
			return
		}
	}
}
