package baidu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	//用户授权跳转
	authorizeUrl = "https://openapi.baidu.com/oauth/2.0/authorize?response_type=code&client_id=__CLIENT_ID__&redirect_uri=__REGISTERED_REDIRECT_URI__&scope=basic,netdisk&display=tv&qrcode=1&force_login=0"
	//token换取access_token链接
	tokenUrl = "https://openapi.baidu.com/oauth/2.0/token?grant_type=authorization_code&code=__CODE__&client_id=__CLIENT_ID__&client_secret=__CLIENT_SECRET__&redirect_uri=__REGISTERED_REDIRECT_URI__"
	//刷新Token有效期链接
	refreshTokenUrl = "https://openapi.baidu.com/oauth/2.0/token?grant_type=refresh_token&refresh_token=__REFRESH_TOKEN__&client_id=__API_KEY__&client_secret=__SECRET_KEY__"
	//获取用户信息
	userInfoUrl = "https://pan.baidu.com/rest/2.0/xpan/nas?method=uinfo&access_token=__ACCESS_TOKEN__"
	//预创建文件
	preCreateFileUrl = "https://pan.baidu.com/rest/2.0/xpan/file?method=precreate&access_token=__ACCESS_TOKEN__"
	//分片上传文件
	uploadFileUrl = "https://d.pcs.baidu.com/rest/2.0/pcs/superfile2?"
	//创建文件
	createFileUrl = "https://pan.baidu.com/rest/2.0/xpan/file?method=create&access_token=__ACCESS_TOKEN__"
)

type Netdisk struct {
	appId     string
	appKey    string
	secretKey string
	signKey   string
	isDebug   bool
	token     *TokenResponse
	user      *UserInfo
	log       *log.Logger
}

func NewNetdisk(appId, appKey, secretKey, signKey string) *Netdisk {
	return &Netdisk{
		appId:     appId,
		appKey:    appKey,
		secretKey: secretKey,
		signKey:   signKey,
		log:       log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile),
	}
}
func (d *Netdisk) IsDebug(isDebug bool) {
	d.isDebug = isDebug
}

//AuthorizeURI 获取用户授权页面.
func (d *Netdisk) AuthorizeURI(registeredUrl string, display string) string {
	urlStr := strings.ReplaceAll(authorizeUrl, "__CLIENT_ID__", d.appKey)
	urlStr = strings.ReplaceAll(urlStr, "__REGISTERED_REDIRECT_URI__", registeredUrl)
	if display != "" {
		urlStr += "&display=" + display
	}
	return urlStr
}

//GetAccessToken 获取access_token值.
func (d *Netdisk) GetAccessToken(code, registeredUrl string) (*TokenResponse, error) {
	d.printf("开始申请access_token:code=%s; registered_url=%s", code, registeredUrl)

	if d.token != nil {
		d.printf("已存在access_token则复用:code=%s; registered_url=%s", code, registeredUrl)

		if d.token.IsExpired() {
			d.printf("access_token已过期则刷新token:code=%s; registered_url=%s", code, registeredUrl)

			if err := d.RefreshToken(true); err == nil {
				return d.token.Clone(), nil
			} else {
				d.printf("刷新token失败:code=%s; registered_url=%s; error=%+v", code, registeredUrl, err)
			}
		} else {
			return d.token.Clone(), nil
		}
	}
	urlStr := strings.ReplaceAll(tokenUrl, "__CODE__", code)
	urlStr = strings.ReplaceAll(urlStr, "__CLIENT_ID__", d.appKey)
	urlStr = strings.ReplaceAll(urlStr, "__CLIENT_SECRET__", d.secretKey)
	urlStr = strings.ReplaceAll(urlStr, "__REGISTERED_REDIRECT_URI__", registeredUrl)

	resp, err := http.Get(urlStr)
	if err != nil {
		d.printf("发起申请token请求失败:code=%s; registered_url=%s; error=%+v", code, registeredUrl, err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.printf("读取申请token结果:code=%s; registered_url=%s; response_body=%s,error=%+v", code, registeredUrl, string(body), err)
		return nil, err
	}
	var errResp ErrorResponse

	err = json.Unmarshal(body, &errResp)
	if err != nil {
		d.printf("解析响应结果失败:code=%s; registered_url=%s; response_body=%s,error=%+v", code, registeredUrl, string(body), err)

		return nil, err
	}
	if errResp.Error != "" {
		d.printf("申请token失败:code=%s; registered_url=%s; response_body=%s,error=%s", code, registeredUrl, string(body), errResp.String())
		return nil, errors.New(errResp.Error + " " + errResp.ErrorDescription)
	}
	tokenResp := errResp.TokenResponse

	tokenResp.CreateAt = time.Now().Unix()
	tokenResp.RefreshTokenCreateAt = time.Now().Unix()
	d.token = (&tokenResp).Clone()

	return &tokenResp, err
}

func (d *Netdisk) SetAccessToken(token *TokenResponse) {
	d.token = token.Clone()
}

func (d *Netdisk) AutoRefreshToken(ctx context.Context) error {
	if d.token == nil {
		return ErrAccessTokenEmpty
	}
	err := d.RefreshToken(false)
	if err != nil {
		return err
	}
	interval := time.Duration(d.token.ExpiresIn-1) * time.Second

	t := time.NewTicker(interval)
	for {
		select {
		case <-t.C:
			err = d.RefreshToken(true)
			if err != nil {
				d.printf("刷新access_token失败 -> %+v", err)
			} else {
				interval = time.Duration(d.token.ExpiresIn-1) * time.Second
			}
			t.Reset(interval)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

//RefreshToken 刷新access_token值.
func (d *Netdisk) RefreshToken(force bool) error {
	if d.token == nil || d.token.IsRefreshTokenExpired() {
		d.printf("access_token未授权或已过期")
		return ErrRefreshTokenExpired
	}
	if !force && !d.token.IsExpired() {
		return nil
	}
	urlStr := strings.ReplaceAll(refreshTokenUrl, "__REFRESH_TOKEN__", d.token.RefreshToken)
	urlStr = strings.ReplaceAll(urlStr, "__CLIENT_ID__", d.appKey)
	urlStr = strings.ReplaceAll(urlStr, "__CLIENT_SECRET__", d.secretKey)

	resp, err := http.Get(urlStr)
	if err != nil {
		d.printf("刷新access_token失败：error:%+v", err)
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.printf("刷新access_token失败：error:%+v", err)
		return err
	}

	var errResp ErrorResponse

	err = json.Unmarshal(body, &errResp)
	if err != nil {
		d.printf("刷新access_token失败：response_body=%s; error:%+v", string(body), err)
		return err
	}
	if errResp.Error == "" {
		return errors.New(errResp.Error + " " + errResp.ErrorDescription)
	}
	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err == nil {
		tokenResp.CreateAt = time.Now().Unix()
		d.token = &tokenResp
	} else {
		d.printf("刷新access_token失败：error:%+v", err)
	}
	return err
}

//UserInfo 获取用户信息.
func (d *Netdisk) UserInfo() (*UserInfo, error) {
	if d.user != nil {
		return d.user.Clone(), nil
	}
	if d.token == nil {
		d.printf("未授权access_token")
		return nil, ErrAccessTokenEmpty
	}
	if d.token.IsExpired() {
		return nil, ErrAccessTokenExpired
	}
	urlStr := strings.ReplaceAll(userInfoUrl, "__ACCESS_TOKEN__", d.token.AccessToken)
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var userinfo UserInfo
	err = json.Unmarshal(body, &userinfo)
	if err != nil {
		return nil, err
	}
	d.user = &userinfo
	return d.user.Clone(), nil
}

//PreCreate 预创建文件.
func (d *Netdisk) PreCreate(uploadFile *PreCreateUploadFileParam) (*PreCreateUploadFile, error) {
	if d.token == nil {
		return nil, ErrAccessTokenEmpty
	}
	if d.token.IsExpired() {
		return nil, ErrAccessTokenExpired
	}
	urlStr := strings.ReplaceAll(preCreateFileUrl, "__ACCESS_TOKEN__", d.token.AccessToken)

	resp, err := http.PostForm(urlStr, uploadFile.Values())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.printf("读取预创建文件结果失败 -> %s", err)
		return nil, err
	}
	var preUploadFile PreCreateUploadFile
	err = json.Unmarshal(body, &preUploadFile)
	if err != nil {
		d.printf("解析预创建文件结果失败 -> %s", string(body))
	}
	return &preUploadFile, err
}

//UploadFile 上传指定路径的文件.
func (d *Netdisk) UploadFile(uploadFile *PreCreateUploadFile, localFile string) ([]SuperFile, error) {
	if d.token == nil {
		return nil, ErrAccessTokenEmpty
	}
	if d.token.IsExpired() {
		return nil, ErrAccessTokenExpired
	}
	f, err := os.Open(localFile)
	if err != nil {
		d.printf("打开文件失败 -> [filename=%s] %+v", localFile, err)
		return nil, err
	}
	defer f.Close()

	superFiles, err := d.UploadFiles(uploadFile, f)
	if err != nil {
		d.printf("上传文件到百度网盘失败 -> %s - %+v", uploadFile, err)
		return nil, fmt.Errorf("upload file fail: filename:%s; error:%w", localFile, errors.Unwrap(err))
	}
	return superFiles, nil
}

//UploadFiles 批量上传文件.
func (d *Netdisk) UploadFiles(uploadFile *PreCreateUploadFile, reader io.Reader) ([]SuperFile, error) {
	if d.token == nil {
		return nil, ErrAccessTokenEmpty
	}
	if d.token.IsExpired() {
		return nil, ErrAccessTokenExpired
	}
	param := SuperFileParam{
		AccessToken: d.token.AccessToken,
		Path:        uploadFile.Path,
		UploadId:    uploadFile.UploadId,
	}
	var superFiles []SuperFile

	b := make([]byte, 4096*1024)

	for i := 0; ; i++ {
		n, err := io.ReadFull(reader, b)
		if err == io.EOF {
			break
		}
		param.PartSeq = i
		urlStr := uploadFileUrl + param.Values().Encode()
		body := bytes.NewBufferString("file=")
		body.Write(b[:n])
		resp, err := http.Post(urlStr, "application/x-www-form-urlencoded", body)
		if err != nil {
			return nil, fmt.Errorf("file index:%d, error:%w", i, err)
		}
		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("file index:%d, error:%w", i, err)
		}
		var superFile SuperFile
		err = json.Unmarshal(respBody, &superFile)
		if err != nil {
			return nil, fmt.Errorf("file index:%d;resp: %s; error:%w", i, string(respBody), err)
		}
		if superFile.ErrNo != 0 {
			return nil, fmt.Errorf("分片上传文件失败 -> resp:%s errcode:%d", string(respBody), superFile.ErrNo)
		}
		d.printf("分片上传成功->[index:%d]  [size:%d] [resp:]", i, n, string(respBody))
		superFiles = append(superFiles, superFile)
	}
	return superFiles, nil
}

func (d *Netdisk) CreateFile(uploadFile *CreateFileParam) (*CreateFile, error) {
	if d.token == nil {
		return nil, ErrAccessTokenEmpty
	}
	if d.token.IsExpired() {
		return nil, ErrAccessTokenExpired
	}
	urlStr := strings.ReplaceAll(createFileUrl, "__ACCESS_TOKEN__", d.token.AccessToken)

	resp, err := http.PostForm(urlStr, uploadFile.Values())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var createFile CreateFile
	err = json.Unmarshal(body, &createFile)

	return &createFile, err
}

func (d *Netdisk) printf(format string, v ...interface{}) {
	if d.isDebug {
		if len(v) == 0 {
			_ = d.log.Output(2, format)
		} else {
			_ = d.log.Output(2, fmt.Sprintf(format, v...))
		}
	}
}
