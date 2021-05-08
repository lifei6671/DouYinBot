package models

import (
	"encoding/json"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

type BaiduUser struct {
	BaiduId              int       `orm:"column(baidu_id);pk;description(百度网盘Id)"`
	BaiduName            string    `orm:"column(baidu_name);size(255);description(百度账号)"`
	NetdiskName          string    `orm:"column(netdisk_name);size(255);description(百度网盘账号)"`
	AvatarUrl            string    `orm:"column(avatar_url);size(2000);null;description(头像地址)"`
	VipType              int       `orm:"column(vip_type);description(会员类型)"`
	AccessToken          string    `orm:"column(access_token);size(500);description(授权码)"`
	ExpiresIn            int64     `orm:"column(expires_in);default(0);description(过期时间，单位秒)"`
	RefreshToken         string    `orm:"column(refresh_token);size(500);description(刷新access_token的token)"`
	Scope                string    `orm:"column(scope);size(1000);description(用户授权的权限)"`
	Created              time.Time `orm:"column(created);auto_now_add;type(datetime);description(创建时间)"`
	Updated              time.Time `orm:"column(updated);auto_now;type(datetime);description(修改时间)"`
	RefreshTokenCreateAt time.Time `orm:"column(refresh_token_create_at);auto_now_add;type(datetime);description(刷新access_token的时间)"`
}

func (b *BaiduUser) TableName() string {
	return "baidu_tokens"
}

func NewBaiduToken() *BaiduUser {
	return &BaiduUser{}
}

func (b *BaiduUser) First(baiduId int) (*BaiduUser, error) {
	o := orm.NewOrm()

	err := o.QueryTable(b.TableName()).Filter("baidu_id", baiduId).One(b)
	return b, err
}

func (b *BaiduUser) Save() (err error) {
	o := orm.NewOrm()
	if o.QueryTable(b.TableName()).Filter("baidu_id", b.BaiduId).Exist() {
		_, err = o.Update(b, "vip_type", "access_token", "expires_in", "refresh_token", "scope", "updated", "refresh_token_create_at")
	} else {
		_, err = o.Insert(b)
	}
	return
}

func (b *BaiduUser) String() string {
	body, _ := json.Marshal(b)
	return string(body)
}
func init() {
	orm.RegisterModel(new(BaiduUser))
}
