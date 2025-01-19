package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"

	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/douyin"
	"github.com/lifei6671/douyinbot/internal/utils"
	"github.com/lifei6671/douyinbot/storage"
)

var (
	ErrNoUserRegister = errors.New("不是用户注册")
	workerNum         = 10
	videoShareChan    = make(chan MediaContent, 100)
	accessKey         = ""
	secretKey         = ""
	bucketName        = ""
	domain            = ""
	savepath          = ""

	fileClient storage.Storage
)

type MediaContent struct {
	Content string
	UserId  string
}

func Push(ctx context.Context, content MediaContent) {
	select {
	case videoShareChan <- content:
	case <-ctx.Done():
	}
}

func Run(ctx context.Context) (err error) {
	if num, err := web.AppConfig.Int("workernumber"); err == nil && num > 0 {
		workerNum = num
	}
	if web.AppConfig.DefaultBool("s3_enable", false) {
		fileClient, err = storage.Factory("cloudflare",
			storage.WithBucketName(web.AppConfig.DefaultString("s3_bucket_name", "")),
			storage.WithAccountID(web.AppConfig.DefaultString("s3_account_id", "")),
			storage.WithAccessKeyID(web.AppConfig.DefaultString("s3_access_key_id", "")),
			storage.WithAccessKeySecret(web.AppConfig.DefaultString("s3_access_key_secret", "")),
			storage.WithEndpoint(web.AppConfig.DefaultString("s3_endpoint", "")),
			storage.WithDomain(web.AppConfig.DefaultString("s3_domain", "")),
		)
		if err != nil {
			return fmt.Errorf("init storage err: %w", err)
		}
	} else if web.AppConfig.DefaultBool("qiniuenable", false) {
		accessKey, err = web.AppConfig.String("qiuniuaccesskey")
		if err != nil {
			logs.Error("获取七牛配置失败 -> [qiuniuaccesskey] - %+v", err)
		}
		secretKey, err = web.AppConfig.String("qiuniusecretkey")
		if err != nil {
			logs.Error("获取七牛配置失败 -> [qiuniusecretkey] - %+v", err)
		}
		bucketName, err = web.AppConfig.String("qiuniubucketname")
		if err != nil {
			logs.Error("获取七牛配置失败 -> [qiuniubucketname] - %+v", err)
		}
		domain, err = web.AppConfig.String("qiniudoamin")
		if err != nil {
			logs.Error("获取七牛配置失败 -> [qiniudoamin] - %+v", err)
			return err
		}
	}
	savepath, err = filepath.Abs(web.AppConfig.DefaultString("auto-save-path", "./"))
	if err != nil {
		logs.Error("获取本地储存目录失败 ->[auto-save-path] %+v", err)
		return err
	}
	for i := 0; i < workerNum; i++ {
		go execute(ctx)
	}
	return nil
}

func execute(ctx context.Context) {
	dy := douyin.NewDouYin(
		web.AppConfig.DefaultString("douyinproxy", ""),
		web.AppConfig.DefaultString("douyinproxyusername", ""),
		web.AppConfig.DefaultString("douyinproxypassword", ""),
	)

	for {
		select {
		case content, ok := <-videoShareChan:
			if !ok {
				return
			}
			logs.Info("开始解析抖音视频任务 -> %s", content)
			video, err := dy.Get(content.Content)
			if err != nil {
				logs.Error("解析抖音视频地址失败 -> 【%s】- %+v", content, err)
				continue
			}
			logs.Info("开始下载抖音视频->%s", video)
			videoPath, err := video.Download(savepath)
			if err != nil {
				logs.Error("下载抖音视频失败 -> 【%s】- %+v", content, err)
				continue
			}
			coverURL := video.OriginCover

			coverPath, err := video.DownloadCover(video.OriginCover, savepath)
			if err == nil {
				coverURL = strings.ReplaceAll("/"+strings.TrimPrefix(coverPath, savepath), "//", "/")
			}
			coverURL = "/cover" + coverURL

			name := strings.TrimPrefix(videoPath, savepath)

			// 将视频上传到S3服务器
			if urlStr, err := uploadFile(ctx, coverPath); err == nil {
				coverURL = urlStr
			}

			// 将封面上传到S3服务器
			if urlStr, err := uploadFile(ctx, videoPath); err == nil {
				video.PlayAddr = urlStr
			}

			user, err := models.NewUser().First(content.UserId)
			if err != nil {
				if errors.Is(err, orm.ErrNoRows) {
					user = models.NewUser()
					user.Id = 1
				} else {
					logs.Error("获取用户失败 -> %s - %+v", content, err)
					continue
				}
			}

			if baseDomain := web.AppConfig.DefaultString("douyin-base-url", ""); baseDomain != "" {
				if uri, err := url.ParseRequestURI(video.OriginCover); err == nil {
					originCover := strings.TrimPrefix(video.OriginCover, "https://")
					originCover = strings.TrimPrefix(originCover, "http://")
					originCover = strings.TrimPrefix(originCover, uri.Host)
					originCover = strings.ReplaceAll(originCover, uri.RawQuery, "")
					video.OriginCover = baseDomain + strings.ReplaceAll(originCover, "//", "/")
				}
			}
			m := models.DouYinVideo{
				UserId:           user.Id,
				Nickname:         video.Author.Nickname,
				Signature:        video.Author.Signature,
				AvatarLarger:     video.Author.AvatarLarger,
				AuthorId:         video.Author.Id,
				AuthorShortId:    video.Author.ShortId,
				VideoRawPlayAddr: video.VideoRawAddr,
				VideoPlayAddr:    video.PlayAddr,
				VideoId:          video.PlayId,
				AwemeId:          video.VideoId,
				VideoCover:       video.OriginCover,
				VideoLocalCover:  coverURL,
				VideoLocalAddr:   "/" + name,
				VideoBackAddr:    string(""),
				Desc:             video.Desc,
				RawLink:          video.RawLink,
			}
			if err := m.Save(); err != nil {
				logs.Error("保存视频到数据库失败 -> 【%s】 - %+v", content, err)
				continue
			}

			if tagErr := models.NewDouYinTag().Create(video.Desc, m.VideoId); tagErr != nil {
				logs.Error("初始视频标签出错 -> %+v", tagErr)
			}

			if len(video.OriginCoverList) > 0 {
				expire, _ := utils.ParseExpireUnix(video.OriginCoverList[0])
				cover := models.DouYinCover{
					VideoId:    m.VideoId,
					Cover:      video.OriginCoverList[0],
					CoverImage: strings.Join(video.OriginCoverList, "|"),
					Expires:    expire,
				}
				if err := cover.Save(m.VideoId); err != nil {
					logs.Error("保存封面失败:【%s】 - %+v", content, err)
				}
			}
			logs.Info("解析抖音视频成功 -> 【%s】- %s", content, m.VideoBackAddr)

		case <-ctx.Done():
			return
		}
	}
}

func Register(content, wechatId string) error {
	if strings.HasPrefix(content, "注册#") {
		items := strings.Split(strings.TrimPrefix(content, "注册#"), "#")
		if len(items) != 3 || items[0] == "" || items[1] == "" || items[2] == "" {
			return errors.New("注册信息格式不正确")
		}
		if !strings.Contains(items[2], "@") {
			return errors.New("邮箱格式不正确")
		}
		user := models.NewUser()
		user.Account = items[0]
		password, err := bcrypt.GenerateFromPassword([]byte(items[1]), bcrypt.DefaultCost)
		if err != nil {
			logs.Error("加密密码失败 -> %+v", err)
			return errors.New("密码格式不正确")
		}
		user.Password = string(password)
		user.WechatId = wechatId
		user.Email = strings.TrimSpace(items[2])
		err = user.Insert()
		if err != nil {
			logs.Error("注册用户失败 -> %+v - %+v", user, err)
			return errors.New("注册用户失败")
		}
		return nil
	}
	return ErrNoUserRegister
}

// 上传文件到S3服务器
func uploadFile(ctx context.Context, filename string) (string, error) {
	if fileClient == nil {
		return filename, nil
	}
	f, err := os.Open(filename)
	if err != nil {
		logs.Error("打开文件失败 -> %s - %+v", filename, err)
		return filename, err
	}
	defer f.Close()

	remoteFilename := strings.TrimPrefix(filename, savepath)

	urlStr, err := fileClient.WriteFile(ctx, f, strings.TrimPrefix(remoteFilename, "/"))
	if err != nil {
		logs.Error("上传文件失败 -> %s - %+v", filename, err)
		return "", err
	}
	return urlStr, nil
}
