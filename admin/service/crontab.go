package service

import (
	"context"
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/douyin"
	"github.com/lifei6671/douyinbot/internal/utils"
	"strings"
	"sync"
	"time"
)

var _cronCh = make(chan string,100)

// RunCron 运行定时任务
func RunCron(ctx context.Context) {
	go func() {
		once := sync.Once{}
		timer := time.NewTicker(time.Second * 12)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				coverList,err := models.NewDouYinCover().GetExpireList()
				if err != nil && !errors.Is(err,orm.ErrNoRows) {
					logs.Error("查询过期列表失败 : %+v",err)
				}
				if err == nil {
					for _,cover := range coverList {
						if cover.Expires > 0 {
							_cronCh <- cover.VideoId
						}
					}
				}
				once.Do(func() {
					timer.Reset(time.Minute * 30)
				})
			case <-ctx.Done():
				return
			}
		}

	}()
	go func() {

		select {
		case videoId, ok := <-_cronCh:
			if !ok {
				return
			}
			err := syncCover(videoId)
			if err != nil {
				logs.Error("更新封面失败: 【%s】 %+v",videoId, err)
			} else {
				logs.Info("更新封面成功: %s", err)
			}

		case <-ctx.Done():
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
		//将状态更新为无效
		_ = models.NewDouYinCover().SetStatus(videoId,1)
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
			} else {
				logs.Info("更新封面成功: %s", videoRecord.VideoCover)
			}
		}
	}
	return nil
}

// AddSyncCover 推送到chan中
func AddSyncCover(videoId string)  {
	_cronCh <- videoId
}