package models

import (
	"encoding/json"
	"github.com/beego/beego/v2/client/orm"
	"time"
)

type User struct {
	Id       int       `orm:"column(id);auto;pk"`
	Account  string    `orm:"column(account);size(255);unique;description(账号)"`
	Password string    `orm:"column(password);size(2000);null;description(密码)" json:"-"`
	Email    string    `orm:"column(email);size(255);unique;description(用户邮箱)"`
	Avatar   string    `orm:"column(avatar);default(/static/avatar/default.jpg);size(1000);" json:"avatar"`
	WechatId string    `orm:"column(wechat_id);size(200);null;description(微信的用户ID)"`
	BaiduId  int       `orm:"column(baidu_id);size(200);null;description(百度网盘用户Id)"`
	Status   int       `orm:"column(status);type(tinyint);default(0);description(用户状态:0=正常/1=禁用/2=删除)" json:"status"`
	Created  time.Time `orm:"column(created);auto_now_add;type(datetime);description(创建时间)"`
	Updated  time.Time `orm:"column(updated);auto_now;type(datetime);description(修改时间)"`
}

func (u *User) TableName() string {
	return "users"
}

func NewUser() *User {
	return &User{}
}

func (u *User) Insert() error {
	o := orm.NewOrm()

	if o.QueryTable(u.TableName()).Filter("account", u.Account).Exist() {
		return ErrUserAccountExist
	}
	if o.QueryTable(u.TableName()).Filter("email", u.Email).Exist() {
		return ErrUsrEmailExist
	}
	if u.WechatId != "" && o.QueryTable(u.TableName()).Filter("wechat_id", u.WechatId).Exist() {
		return ErrUserWechatIdExist
	}
	id, err := o.Insert(u)
	if err != nil {
		return err
	}
	u.Id = int(id)
	return nil
}

func (u *User) Update(cols ...string) error {
	o := orm.NewOrm()
	_, err := o.Update(u, cols...)
	return err
}

func (u *User) First(account string) (*User, error) {
	o := orm.NewOrm()
	cond := orm.NewCondition().And("account", account).
		Or("email", account).
		Or("wechat_id", account).
		Or("baidu_id", account)

	err := o.QueryTable(u.TableName()).SetCond(cond).One(u)

	return u, err
}

func (u *User) FirstByWechatId(id string) (*User, error) {
	err := orm.NewOrm().QueryTable(u.TableName()).Filter("wechat_id", id).One(u)

	return u, err
}

func (u *User) ExistByWechatId(id string) bool {
	return orm.NewOrm().QueryTable(u.TableName()).Filter("wechat_id", id).Exist()
}
func (u *User) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}

func init() {
	// 需要在init中注册定义的model
	orm.RegisterModel(new(User))
}
