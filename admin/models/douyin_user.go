package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type DouYinUser struct {
	Id           int       `orm:"column(id);auto;pk"`
	Nickname     string    `orm:"column(nickname);size(100); description(作者昵称)"`
	Signature    string    `orm:"column(signature);size(255);null;description(作者信息)"`
	AvatarLarger string    `orm:"column(avatar_larger);size(2000);null;description(作者头像)"`
	AvatarCDNURL string    `orm:"column(avatar_cdn_url);size(2000);null;description(作者头像)"`
	HashValue    string    `orm:"column(hash_value);index;size(64);null;description(作者头像)"`
	AuthorId     string    `orm:"column(author_id);size(20);null;description(作者长ID)"`
	Created      time.Time `orm:"auto_now_add;type(datetime);description(创建时间)"`
}

func (d *DouYinUser) TableName() string {
	return "douyin_user"
}

// TableUnique 多字段唯一键
func (d *DouYinUser) TableUnique() [][]string {
	return [][]string{
		{"AuthorId"},
	}
}

func NewDouYinUser() *DouYinUser {
	return &DouYinUser{}
}

func (d *DouYinUser) GetById(authorId string) (*DouYinUser, error) {
	err := orm.NewOrm().QueryTable(d.TableName()).Filter("author_id", authorId).One(d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DouYinUser) Create() (int, error) {
	id, err := orm.NewOrm().Insert(d)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (d *DouYinUser) Update() error {
	_, err := orm.NewOrm().Update(d)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	// 需要在init中注册定义的model
	orm.RegisterModel(new(DouYinUser))
}
