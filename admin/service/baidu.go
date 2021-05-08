package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin/models"
	"github.com/lifei6671/douyinbot/baidu"
	"time"
)

var (
	baiduAppId     = web.AppConfig.DefaultString("baiduappid", "")
	baiduAppKey    = web.AppConfig.DefaultString("baiduappkey", "")
	baiduSecretKey = web.AppConfig.DefaultString("baidusecretkey", "")
	baiduSignKey   = web.AppConfig.DefaultString("baidusignkey", "")
	baiduCache     = cache.NewMemoryCache()
)

func uploadBaiduNetdisk(ctx context.Context, baiduId int, filename string, remoteName string) (*baidu.CreateFile, error) {
	key := fmt.Sprintf("baidu::%d", baiduId)
	val, _ := baiduCache.Get(ctx, key)
	bd, ok := val.(*baidu.Netdisk)
	if !ok || bd == nil {
		token, err := models.NewBaiduToken().First(baiduId)
		if err != nil {
			return nil, fmt.Errorf("用户未绑定百度网盘：[baiduid=%d] - %w", baiduId, err)
		}
		bd = baidu.NewNetdisk(baiduAppId, baiduAppKey, baiduSecretKey, baiduSignKey)
		bd.SetAccessToken(&baidu.TokenResponse{
			AccessToken:          token.AccessToken,
			ExpiresIn:            token.ExpiresIn,
			RefreshToken:         token.RefreshToken,
			Scope:                token.Scope,
			CreateAt:             token.Created.Unix(),
			RefreshTokenCreateAt: token.RefreshTokenCreateAt.Unix(),
		})
		bd.IsDebug(true)
		_ = bd.RefreshToken(false)

		_ = baiduCache.Put(ctx, key, bd, time.Duration(token.ExpiresIn)*time.Second)
	} else {
		_ = bd.RefreshToken(false)
	}

	uploadFile, err := baidu.NewPreCreateUploadFileParam(filename, remoteName)
	if err != nil {
		logs.Error("预创建文件失败 -> [filename=%s] ; %+v", remoteName, err)
		return nil, fmt.Errorf("预创建文件失败 -> [filename=%s] ; %w", remoteName, err)
	}
	logs.Info("开始预创建文件 ->%s", uploadFile)
	preUploadFile, err := bd.PreCreate(uploadFile)
	if err != nil {
		logs.Error("预创建文件失败 -> [filename=%s] ; %+v", remoteName, err)
		return nil, fmt.Errorf("预创建文件失败 -> [filename=%s] ; %w", remoteName, err)
	}
	logs.Info("开始分片上传文件 -> %s", preUploadFile)

	superFiles, err := bd.UploadFile(preUploadFile, filename)
	if err != nil {
		logs.Error("创建文件失败 -> [filename=%s] ; %+v", remoteName, err)
		return nil, fmt.Errorf("创建文件失败 -> [filename=%s] ; %w", remoteName, err)
	}
	b, _ := json.Marshal(&superFiles)
	logs.Info("分片上传成功 -> %s", string(b))

	param := baidu.NewCreateFileParam(remoteName, uploadFile.Size, false)
	param.BlockList = make([]string, len(superFiles))
	param.UploadId = preUploadFile.UploadId
	//文件命名策略，默认1
	//0 为不重命名，返回冲突
	//1 为只要path冲突即重命名
	//2 为path冲突且block_list不同才重命名
	//3 为覆盖
	param.RType = 3
	param.Path = "/apps/DouYinBot" + param.Path

	for i, f := range superFiles {
		param.BlockList[i] = f.Md5
	}
	logs.Info("最终合并文件 -> %s", param)
	createFile, err := bd.CreateFile(param)
	if err != nil {
		logs.Error("创建文件失败 -> [filename=%s] ; %+v", remoteName, err)
		return nil, fmt.Errorf("创建文件失败 -> [filename=%s] ; %w", remoteName, err)
	}
	logs.Info("百度网盘上传成功 -> %s", createFile)
	return createFile, nil
}
