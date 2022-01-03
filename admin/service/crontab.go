package service

import (
	"context"
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/jasonlvhit/gocron"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/douyin"
	"github.com/lifei6671/douyinbot/internal/utils"
	"strings"
	"time"
)

// RunCron 运行定时任务
func RunCron(ctx context.Context) {
	ch := gocron.Start()
	go func() {
		coverList,err := models.NewDouYinCover().GetExpireList()
		if err != nil && !errors.Is(err,orm.ErrNoRows) {
			logs.Error("查询过期列表失败 : %+v",err)
		}
		if err == nil {
			for _,cover := range coverList {
				if cover.Expires > 0 {
					t := time.Unix(int64(cover.Expires), 0)
					err = gocron.Every(1).From(&t).Do(syncCover(cover.VideoId))
					if err != nil {
						logs.Error("加入封面过期队列失败:【%s】- %+v", cover.VideoId, err)
						continue
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			ch <- true
			return
		}
	}()
}

// syncCover 同步过期的封面
func syncCover(videoId string) error {
	videoRecord, err := models.NewDouYinVideo().FirstByVideoId(videoId)
	if err != nil {
		return err
	}
	logs.Info("开始解析抖音视频任务 -> %s", videoRecord.RawLink)
	dy := douyin.NewDouYin()

	video, err := dy.Get(videoRecord.RawLink)
	if err != nil {
		logs.Error("解析抖音视频地址失败 -> 【%s】- %+v", videoRecord.RawLink, err)
		return err
	}
	if len(video.OriginCoverList) > 0 {
		expire, _ := utils.ParseExpireUnix(video.OriginCoverList[0])
		cover := models.DouYinCover{
			VideoId:    videoRecord.VideoId,
			Cover:      video.OriginCoverList[0],
			CoverImage: strings.Join(video.OriginCoverList, "|"),
			Expires:    expire,
		}
		if err := cover.Save(videoRecord.VideoId); err != nil {
			logs.Error("保存封面失败: %+v", err)
		} else {
			videoRecord.VideoCover = video.OriginCover
			if err := videoRecord.Save(); err != nil {
				logs.Error("保存默认封面:【%s】- %+v", videoRecord.RawLink, err)
				return err
			}
			if expire > 0 {
				t := time.Unix(int64(expire), 0)
				//加入过期更新队列
				err = gocron.Every(1).From(&t).Do(syncCover(videoRecord.VideoId))
				if err != nil {
					logs.Error("加入封面过期队列失败:【%s】- %+v", videoRecord.RawLink, err)
					return err
				}
			}
		}
	}
	return nil
}
