package models

import (
	"errors"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

var PageSize = 18

type DouYinVideo struct {
	Id               int       `orm:"column(id);auto;pk"`
	UserId           int       `orm:"column(user_id);index;description(所属用户)"`
	Nickname         string    `orm:"column(nickname);size(100); description(作者昵称)"`
	Signature        string    `orm:"column(signature);size(255);null;description(作者信息)"`
	AvatarLarger     string    `orm:"column(avatar_larger);size(2000);null;description(作者头像)"`
	AuthorId         string    `orm:"column(author_id);size(20);null;description(作者长ID)"`
	AuthorShortId    string    `orm:"column(author_short_id);size(10);null;description(作者短ID)"`
	VideoRawPlayAddr string    `orm:"column(video_raw_play_addr);size(2000);description(原视频地址)"`
	VideoPlayAddr    string    `orm:"column(video_play_addr);size(2000);description(视频原播放地址)"`
	VideoId          string    `orm:"column(video_id);size(255);unique;description(视频唯一ID)"`
	AwemeId          string    `orm:"column(aweme_id);size(255);description(原始awemeid)"`
	VideoCover       string    `orm:"column(video_cover);size(2000);null;description(视频封面)"`
	VideoLocalCover  string    `orm:"column(video_local_cover);size(2000);description(本地备份封面)"`
	VideoLocalAddr   string    `orm:"column(video_local_addr);size(2000);description(本地路径)"`
	VideoBackAddr    string    `orm:"column(video_back_addr);size(2000);null;description(备份的地址)"`
	Desc             string    `orm:"column(desc);size(1000);null;description(视频描述)"`
	RawLink          string    `orm:"column(raw_link);size(255);default('');description(原始分享内容)"`
	Created          time.Time `orm:"auto_now_add;type(datetime);description(创建时间)"`
}

func (d *DouYinVideo) TableName() string {
	return "douyin_video"
}

func NewDouYinVideo() *DouYinVideo {
	return &DouYinVideo{}
}

func (d *DouYinVideo) GetList(pageIndex int, authorId int) (list []DouYinVideo, total int, err error) {
	o := orm.NewOrm()
	offset := (pageIndex - 1) * PageSize
	query := o.QueryTable(d.TableName()).OrderBy("-id")
	if authorId > 0 {
		query = query.Filter("author_id", authorId)
	}
	count, err := query.Count()
	total = int(count)

	_, err = query.Offset(offset).Limit(PageSize).All(&list)

	return
}

func (d *DouYinVideo) Save() error {
	o := orm.NewOrm()

	var video DouYinVideo

	err := o.QueryTable(d.TableName()).Filter("video_id", d.VideoId).One(&video)
	if errors.Is(err, orm.ErrNoRows) {
		_, err = o.Insert(d)
	} else if err == nil {
		d.Id = video.Id
		_, err = o.Update(d)
	}

	return err
}

func (d *DouYinVideo) FirstByVideoId(videoId string) (*DouYinVideo, error) {
	o := orm.NewOrm()
	err := o.QueryTable(d.TableName()).Filter("video_id", videoId).One(d)
	if err != nil && !errors.Is(err, orm.ErrNoRows) {
		logs.Error("查询视频失败 -> video_id=%s ; error=%+v", videoId, err)
		return nil, err
	}
	return d, nil
}

// Next 查询指定视频的下一条视频
func (d *DouYinVideo) Next(videoId string) (*DouYinVideo, error) {
	o := orm.NewOrm()

	var current DouYinVideo
	err := o.QueryTable(d.TableName()).Filter("video_id", videoId).One(&current)
	if err != nil {
		return nil, err
	}
	// 查询下一条
	var nextVideo DouYinVideo
	errNext := o.QueryTable(d.TableName()).
		Filter("id__gt", current.Id).
		OrderBy("id").
		Limit(1).
		One(&nextVideo)
	if errNext != nil {
		return nil, errNext
	}
	return &nextVideo, nil
}

// Prev 查询指定视频的上一条视频
func (d *DouYinVideo) Prev(videoId string) (*DouYinVideo, error) {
	o := orm.NewOrm()

	var current DouYinVideo
	err := o.QueryTable(d.TableName()).Filter("video_id", videoId).One(&current)
	if err != nil {
		return nil, err
	}
	// 查询下一条
	var prevVideo DouYinVideo
	errPrev := o.QueryTable(d.TableName()).
		Filter("id__lt", current.Id).
		OrderBy("-id"). // 降序排列后取第一条
		Limit(1).
		One(&prevVideo)
	if errPrev != nil {
		return nil, errPrev
	}
	return &prevVideo, nil
}

func init() {
	// 需要在init中注册定义的model
	orm.RegisterModel(new(DouYinVideo))
}
