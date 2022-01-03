package models

import (
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"golang.org/x/exp/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type DouYinCover struct {
	Id         int       `orm:"column(id);auto;pk"`
	VideoId    string    `orm:"column(video_id);size(255);unique;description(视频唯一ID)"`
	Cover      string    `orm:"column(cover);size(255);description(第一张封面)"`
	CoverImage string    `orm:"column(cover_image);size(2000);description(封面地址)"`
	Expires    int       `orm:"column(expires);description(封面有效期)"`
	Created    time.Time `orm:"auto_now_add;type(datetime);description(创建时间)"`
}

func NewDouYinCover() *DouYinCover {
	return new(DouYinCover)
}
func (d *DouYinCover) TableName() string {
	return "douyin_cover"
}

// Save 更新或插入封面信息
func (d *DouYinCover) Save(videoId string) error {
	if len(d.CoverImage) == 0 {
		return nil
	}
	o := orm.NewOrm()

	if d.Expires == 0 {
		uri, err := url.ParseRequestURI(d.Cover)
		if err != nil {
			return err
		}
		if v := uri.Query().Get("x-expires"); v != "" {
			expire, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			d.Expires = expire
		}
	}
	var cover DouYinCover
	err := o.QueryTable(d.TableName()).Filter("video_id", videoId).One(&cover)
	if err == orm.ErrNoRows {
		_, err = o.Insert(d)
	} else if err == nil {
		d.Id = cover.Id
		_, err = o.Update(d)
	}

	return err
}

// CoverFirst 获取视频的地址
func (d *DouYinCover) CoverFirst(videoId string) (string, error) {
	o := orm.NewOrm()
	var cover *DouYinCover
	err := o.QueryTable(d.TableName()).Filter("video_id", videoId).One(&cover)
	if err != nil {
		return "", err
	}
	covers := strings.Split(cover.CoverImage, "|")
	if len(covers) > 0 {
		return covers[rand.Intn(len(covers))], nil
	}
	return "", orm.ErrNoRows
}

// GetExpireList 获取临近过期的封面
func (d *DouYinCover) GetExpireList() ([]DouYinCover, error) {
	o := orm.NewOrm()
	var covers []DouYinCover

	_, err := o.QueryTable(d.TableName()).Filter("expires__lte", time.Now().Unix() - 600).All(&covers)
	if err != nil {
		logs.Error("查询过期封面失败： %+v", err)
	}
	return covers, err
}

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
	// 需要在init中注册定义的model
	orm.RegisterModel(new(DouYinCover))
}
