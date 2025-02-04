package admin

import (
	context2 "context"
	"embed"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/lifei6671/fink-download/fink"

	"github.com/lifei6671/douyinbot/admin/controllers"
	_ "github.com/lifei6671/douyinbot/admin/routers"
	"github.com/lifei6671/douyinbot/admin/service"
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
			return fmt.Errorf("load config file failed: %s - %w", configFile, err)
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
		if b, err := os.ReadFile(filepath.Join(web.WorkPath, "static/video/default.mp4")); err == nil {
			controllers.SetDefaultVideoContent(b)
		}
	}

	web.Get("/robots.txt", func(ctx *context.Context) {
		if configFile != "" {
			robotsPath := filepath.Join(filepath.Dir(configFile), "robots.txt")

			b, err := os.ReadFile(robotsPath)
			if err != nil {
				ctx.Output.SetStatus(http.StatusNotFound)
				return
			}
			ctx.Output.Header("X-Content-Type-Options", "nosniff")
			err = ctx.Output.Body(b)
			if err != nil {
				logs.Error("写入数据到客户端失败 -> %+v", err)
			}
		}
		ctx.Output.SetStatus(http.StatusNotFound)
		return
	})

	web.Get("/static/*.*", func(ctx *context.Context) {
		var b []byte
		var err error
		if web.BConfig.RunMode == web.PROD {
			//读取文件
			b, err = Assets.ReadFile(strings.TrimPrefix(ctx.Request.URL.Path, "/"))
		} else {
			b, err = os.ReadFile(filepath.Join(web.WorkPath, ctx.Request.URL.Path))
		}
		if err != nil {
			logs.Error("文件不存在 -> %s", ctx.Request.URL.Path)
			ctx.Output.SetStatus(404)
			return
		}
		//解析文件类型
		contentType := mime.TypeByExtension(filepath.Ext(ctx.Request.URL.Path))
		if contentType != "" {
			ctx.Output.Header("Content-Type", contentType)
			ctx.Output.Header("X-Content-Type-Options", "nosniff")
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
		ctx.Output.Header("Cache-Control", fmt.Sprintf("max-age=%d, s-maxage=%d", 3600*30*24, 3600*30*24))
		ctx.Output.Header("Cloudflare-CDN-Cache-Control", "max-age=14400")
		ctx.Output.Header("Last-Modified", RunTime.Format(modifiedFormat))

		err = ctx.Output.Body(b)
		if err != nil {
			logs.Error("写入数据到客户端失败 -> %+v", err)
		}
	})

	savePath, err := web.AppConfig.String("auto-save-path")
	if err == nil {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			if err := os.MkdirAll(savePath, 0755); err != nil {
				return fmt.Errorf("mkdir fail ->%s - %w", savePath, err)
			}
		}
		//web.SetStaticPath("/video", savePath)
	}
	web.Get("/cover/*.*", func(ctx *context.Context) {
		filename := filepath.Join(savePath, strings.TrimPrefix(ctx.Request.RequestURI, "/cover"))
		b, err := os.ReadFile(filename)
		if err != nil {
			logs.Error("文件不存在 -> %s", ctx.Request.RequestURI)
			ctx.Output.SetStatus(404)
			return
		}
		//解析文件类型
		contentType := mime.TypeByExtension(filepath.Ext(filename))
		if contentType != "" {
			ctx.Output.Header("Content-Type", contentType)
			ctx.Output.Header("X-Content-Type-Options", "nosniff")
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
		ctx.Output.Header("Cache-Control", fmt.Sprintf("max-age=%d, s-maxage=%d", 3600*30*24, 3600*30*24))
		ctx.Output.Header("Cloudflare-CDN-Cache-Control", "max-age=14400")

		err = ctx.Output.Body(b)
		if err != nil {
			logs.Error("写入数据到客户端失败 -> %+v", err)
		}
	})
	imagePath, err := web.AppConfig.String("image-save-path")

	if err == nil {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			if err := os.MkdirAll(savePath, 0755); err != nil {
				return fmt.Errorf("mk dir err %s - %+w", savePath, err)
			}
		}
		go func() {
			if err := fink.Run(context2.Background(), imagePath); err != nil {
				panic(fmt.Errorf("create image path err %s - %w", savePath, err))
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
