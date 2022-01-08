package admin

import (
	context2 "context"
	"embed"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/lifei6671/douyinbot/admin/controllers"
	_ "github.com/lifei6671/douyinbot/admin/routers"
	"github.com/lifei6671/douyinbot/admin/service"
	"github.com/lifei6671/fink-download/fink"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed views static
var Assets embed.FS
var RunTime = time.Now()
var modifiedFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func Run(addr string, configFile string) error {
	if configFile != "" {
		logs.Info("从文件加载配置文件 -> %s", configFile)
		err := web.LoadAppConfig("ini", configFile)
		if err != nil {
			logs.Error("加载配置文件失败 -> %s - %+v", configFile, err)
			return err
		}
	}

	if web.BConfig.RunMode == web.PROD {
		web.SetTemplateFSFunc(func() http.FileSystem {
			return http.FS(Assets)
		})
		web.SetViewsPath("views")
		if b, err := Assets.ReadFile("static/video/default.mp4"); err == nil {
			controllers.SetDefaultVideoContent(b)
		}
	} else {
		web.SetViewsPath(filepath.Join(web.WorkPath, "views"))
		if b, err := ioutil.ReadFile(filepath.Join(web.WorkPath, "static/video/default.mp4")); err == nil {
			controllers.SetDefaultVideoContent(b)
		}
	}
	web.Get("/static/*.*", func(ctx *context.Context) {
		var b []byte
		var err error
		if web.BConfig.RunMode == web.PROD {
			//读取文件
			b, err = Assets.ReadFile(strings.TrimPrefix(ctx.Request.RequestURI, "/"))
		} else {
			b, err = os.ReadFile(filepath.Join(web.WorkPath, ctx.Request.RequestURI))
		}
		if err != nil {
			logs.Error("文件不存在 -> %s", ctx.Request.RequestURI)
			ctx.Output.SetStatus(404)
			return
		}
		//解析文件类型
		contentType := mime.TypeByExtension(filepath.Ext(ctx.Request.RequestURI))
		if contentType != "" {
			ctx.Output.Header("Content-Type", contentType)
		}
		//解析客户端文件版本
		modified := ctx.Request.Header.Get("If-Modified-Since")
		if last, err := time.Parse(modifiedFormat, modified); err == nil {
			if RunTime.Before(last) {
				ctx.Output.SetStatus(304)
				return
			}
		}
		//写入缓冲时间
		ctx.Output.Header("Cache-Control", fmt.Sprintf("max-age=%d", 3600*30*24))
		ctx.Output.Header("Last-Modified", RunTime.Format(modifiedFormat))

		err = ctx.Output.Body(b)
		if err != nil {
			logs.Error("写入数据到客户端失败 -> %+v", err)
		}
	})

	savePath, err := web.AppConfig.String("auto-save-path")
	if err == nil {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			if err := os.MkdirAll(savePath, 0655); err != nil {
				return err
			}
		}
		//web.SetStaticPath("/video", savePath)
	}
	imagePath, err := web.AppConfig.String("image-save-path")

	if err == nil {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			if err := os.MkdirAll(savePath, 0655); err != nil {
				return err
			}
		}
		go func() {
			if err := fink.Run(context2.Background(), imagePath); err != nil {
				panic(err)
			}
		}()
	}

	if err := service.Run(context2.Background()); err != nil {
		return err
	}

	go service.RunCron(context2.Background())

	web.Run(addr)
	return nil
}
