package service

import (
	"context"
	"errors"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"

	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/internal/utils"
)

var (
	ErrDownloadCover = errors.New("download cover error")
)
var (
	_downloadQueue = make(chan models.DouYinVideo, 100)
)

func PushDownloadQueue(video models.DouYinVideo) {
	_downloadQueue <- video
}

func ExecDownloadQueue(videoModel models.DouYinVideo) {
	log.Println(videoModel.VideoId, videoModel.VideoLocalCover)
	if videoModel.VideoLocalCover == "/cover" || videoModel.VideoLocalCover == "/cover/" {
		avatarPath, err := utils.DownloadCover(videoModel.AuthorId, videoModel.VideoCover, savepath)
		if err != nil {
			var uri *url.URL
			uri, err = url.ParseRequestURI(videoModel.VideoCover)
			if err != nil {
				logs.Error("解析封面文件失败: url[%s] filename[%s] %+v", videoModel.VideoCover, err)
				return
			}
			if !strings.HasPrefix(uri.Host, "p5-ipv6") {
				uri.Host = "p5-ipv6.douyinpic.com"
			}
			avatarPath, err = utils.DownloadCover(videoModel.AuthorId, uri.String(), savepath)

			if err == nil {
				videoModel.VideoCover = uri.String()
			}
			videoModel.VideoLocalCover = "/static/images/default.jpg"
		}
		if err != nil {
			logs.Error("下载视频封面失败: url[%s] filename[%s] %+v", videoModel.VideoCover, err)
		}
		if err == nil {
			videoModel.VideoLocalCover = "/cover" + strings.ReplaceAll("/"+strings.TrimPrefix(avatarPath, savepath), "//", "/")

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			// 将封面上传到S3服务器
			if urlStr, err := uploadFile(ctx, avatarPath); err == nil {
				videoModel.VideoLocalCover = urlStr
			}
		}
	} else if strings.HasPrefix(videoModel.VideoLocalCover, "/cover") {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		coverPath := filepath.Join(savepath, videoModel.VideoLocalCover)
		if utils.FileExists(coverPath) {
			// 将封面上传到S3服务器
			if urlStr, err := uploadFile(ctx, coverPath); err == nil {
				videoModel.VideoLocalCover = urlStr
			}
		}
	}

	if err := videoModel.Save(); err != nil {
		logs.Error("下载视频封面失败 -> [video_id=%s] %+v", videoModel.VideoId, err)
	}
}
