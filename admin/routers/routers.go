package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/controllers"
)

func init() {
	web.Router("/douyin", &controllers.HomeController{}, "get,post:Index")
	web.Router("/douyin/download", &controllers.HomeController{}, "get:Download")
}
