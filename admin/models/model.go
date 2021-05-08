package models

import (
	"errors"
	"github.com/beego/beego/v2/client/orm"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrUserAccountExist  = errors.New("用户账号已存在")
	ErrUsrEmailExist     = errors.New("用户邮箱已存在")
	ErrUserWechatIdExist = errors.New("微信已绑定其他账号")
)

func Init(dataSource string) error {
	filename, err := filepath.Abs(dataSource)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0655); err != nil {
			return err
		}
	}

	// 参数1        数据库的别名，用来在 ORM 中切换数据库使用
	// 参数2        driverName
	// 参数3        对应的链接字符串
	filename = filename + "?_loc=" + time.Local.String()
	if err := orm.RegisterDataBase("default", "sqlite3", filename); err != nil {
		return err
	}
	err = orm.RunSyncdb("default", false, true)
	return err
}
