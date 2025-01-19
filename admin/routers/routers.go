package routers

import (
	"github.com/beego/beego/v2/server/web"

	"github.com/lifei6671/douyinbot/admin/controllers"
)

func init() {
	web.Router("/", &controllers.IndexController{}, "get:Index")
	web.Router("/page/:page:int.html", &controllers.IndexController{}, "get:Index")
	web.Router("/page/:author_id:int_:page:int.html", &controllers.IndexController{}, "get:List")
	web.Router("/douyin", &controllers.HomeController{}, "get,post:Index")
	web.Router("/douyin/download", &controllers.HomeController{}, "get:Download")
	web.Router("/tag/:tag_id:int_:page:int.html", &controllers.TagController{}, "get:Index")

	web.Router("/wechat", &controllers.WeiXinController{}, "get:Index")
	web.Router("/wechat", &controllers.WeiXinController{}, "post:Dispatch")

	web.Router("/baidu/authorize", &controllers.BaiduController{}, "get:Index")
	web.Router("/baidu", &controllers.BaiduController{}, "get:Authorize")

	web.Router("/video/local/play", &controllers.VideoController{}, "get:Index")
	web.Router("/video/remote/play", &controllers.VideoController{}, "get:Play")
}
