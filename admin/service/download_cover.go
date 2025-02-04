package service

import (
	"context"
	"errors"
	"log"
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
			videoModel.VideoLocalCover = "/static/images/default.jpg"
		} else {
			videoModel.VideoLocalCover = "/cover" + strings.ReplaceAll("/"+strings.TrimPrefix(avatarPath, savepath), "//", "/")

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			// 将封面上传到S3服务器
			if urlStr, err := uploadFile(ctx, avatarPath); err == nil {
				videoModel.VideoLocalCover = urlStr
			}
		}
		if err := videoModel.Save(); err != nil {
			logs.Error("下载视频封面失败 -> [video_id=%s] %+v", videoModel.VideoId, err)
		}
	}
}
