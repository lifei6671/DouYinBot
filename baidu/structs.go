package baidu

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ErrRefreshTokenExpired = errors.New("refresh_token expired")
	ErrAccessTokenEmpty    = errors.New("user not authorized ")
	ErrAccessTokenExpired  = errors.New("access token expired")
)

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *ErrorResponse) String() string {
	return fmt.Sprintf("error: %s;error_description: %s", e.Error, e.ErrorDescription)
}

type TokenResponse struct {
	AccessToken          string `json:"access_token"`
	ExpiresIn            int64  `json:"expires_in"`
	RefreshToken         string `json:"refresh_token"`
	Scope                string `json:"scope"`
	SessionKey           string `json:"session_key"`
	SessionSecret        string `json:"session_secret"`
	CreateAt             int64  `json:"-"`
	RefreshTokenCreateAt int64  `json:"-"`
}

func (t *TokenResponse) Clone() *TokenResponse {
	return &TokenResponse{
		AccessToken:          t.AccessToken,
		ExpiresIn:            t.ExpiresIn,
		RefreshToken:         t.RefreshToken,
		Scope:                t.Scope,
		SessionKey:           t.SessionKey,
		SessionSecret:        t.SessionSecret,
		CreateAt:             t.CreateAt,
		RefreshTokenCreateAt: t.RefreshTokenCreateAt,
	}
}
func (t *TokenResponse) IsExpired() bool {
	return time.Now().Unix() >= t.CreateAt+t.ExpiresIn
}

func (t *TokenResponse) IsRefreshTokenExpired() bool {
	return time.Now().AddDate(-10, 0, 0).Unix() >= t.RefreshTokenCreateAt
}

type UserInfo struct {
	ErrNo       int    `json:"errno"`
	ErrMsg      string `json:"errmsg"`
	BaiduName   string `json:"baidu_name"`
	NetdiskName string `json:"netdisk_name"`
	AvatarUrl   string `json:"avatar_url"`
	VipType     int    `json:"vip_type"`
	UserId      int    `json:"uk"`
}

func (u *UserInfo) Clone() *UserInfo {
	return &UserInfo{
		ErrNo:       u.ErrNo,
		ErrMsg:      u.ErrMsg,
		BaiduName:   u.BaiduName,
		NetdiskName: u.NetdiskName,
		AvatarUrl:   u.AvatarUrl,
		VipType:     u.VipType,
		UserId:      u.UserId,
	}
}
func (u *UserInfo) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}

type PreCreateUploadFileParam struct {
	Path       string   `json:"path"`
	Size       int      `json:"size"`
	IsDir      bool     `json:"is_dir"`
	AutoInit   int      `json:"autoinit"`
	RType      int      `json:"rtype"`
	UploadId   string   `json:"uploadid"`
	BlockList  []string `json:"block_list"`
	ContentMD5 string   `json:"content_md5"`
	SliceMD5   string   `json:"slice_md5"`
	LocalCTime int64    `json:"local_ctime"`
	LocalMTime int64    `json:"local_mtime"`
}

func NewPreCreateUploadFileParam(filename string, path string) (*PreCreateUploadFileParam, error) {
	reader, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	info, err := reader.Stat()
	if err != nil {
		return nil, err
	}
	b := make([]byte, 4096)

	blockList := make([]string, 0)
	for i := 0; ; i++ {
		b = b[:0]
		n, err := io.ReadFull(reader, b)
		if err == io.EOF {
			break
		}
		has := md5.Sum(b[:n])
		md5str1 := fmt.Sprintf("%x", has)
		blockList = append(blockList, md5str1)
	}
	sliceBody := make([]byte, 256*1024)
	n, _ := io.ReadFull(reader, sliceBody)
	h := md5.Sum(sliceBody[:n])

	return &PreCreateUploadFileParam{
		Path:       path,
		Size:       int(info.Size()),
		IsDir:      false,
		AutoInit:   1,
		RType:      3,
		BlockList:  blockList,
		UploadId:   blockList[0],
		SliceMD5:   fmt.Sprintf("%x", h),
		LocalCTime: info.ModTime().Unix(),
		LocalMTime: time.Now().Unix(),
	}, nil
}

func (u *PreCreateUploadFileParam) Values() url.Values {
	values := url.Values{}
	values.Add("path", u.Path)
	values.Add("size", fmt.Sprintf("%d", u.Size))
	if u.IsDir {
		values.Add("is_dir", "1")
	} else {
		values.Add("is_dir", "0")
	}
	values.Add("autoinit", "1")
	if u.RType > 0 {
		values.Add("rtype", fmt.Sprintf("%d", u.RType))
	} else {
		values.Add("rtype", "0")
	}
	if u.UploadId != "" {
		values.Add("uploadid", u.UploadId)
	}
	if u.BlockList != nil && len(u.BlockList) > 0 {
		values.Add("block_list", fmt.Sprintf("[\"%s\"]", strings.Join(u.BlockList, "\",\"")))
	}
	if u.ContentMD5 != "" {
		values.Add("content-md5", u.ContentMD5)
	}
	if u.SliceMD5 != "" {
		values.Add("slice-md5", u.SliceMD5)
	}
	if u.LocalMTime > 0 {
		values.Add("local_ctime", fmt.Sprintf("%d", u.LocalCTime))
	}
	if u.LocalMTime > 0 {
		values.Add("local_mtime", fmt.Sprintf("%d", u.LocalMTime))
	}

	return values
}

func (u *PreCreateUploadFileParam) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}

type PreCreateUploadFile struct {
	ErrNo      int            `json:"errno"`
	Path       string         `json:"path"`
	UploadId   string         `json:"uploadid"`
	ReturnType int            `json:"return_type"`
	BlockList  []string       `json:"block_list"`
	Info       UploadFileInfo `json:"info,omitempty"`
}

func (u *PreCreateUploadFile) String() string {
	b, _ := json.Marshal(u)
	return string(b)
}

type UploadFileInfo struct {
	Size     int    `json:"size"`
	Category int    `json:"category"`
	IsDir    int    `json:"is_dir"`
	Path     string `json:"path"`
	FsId     int64  `json:"fs_id"`
	MD5      string `json:"md5"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
}

type SuperFileParam struct {
	AccessToken string
	Method      string
	Type        string
	Path        string
	UploadId    string
	PartSeq     int
}

func (s *SuperFileParam) Values() url.Values {
	values := url.Values{}
	values.Add("access_token", s.AccessToken)
	values.Add("method", "upload")
	values.Add("type", "tmpfile")
	values.Add("path", s.Path)
	values.Add("uploadid", s.UploadId)
	values.Add("partseq", fmt.Sprintf("%d", s.PartSeq))
	return values
}

type SuperFile struct {
	ErrNo     int    `json:"err_no"`
	Md5       string `json:"md5"`
	RequestId string `json:"request_id"`
}

type CreateFileParam struct {
	Path       string   `json:"path"`
	Size       int      `json:"size"`
	IsDir      bool     `json:"isdir"`
	RType      int      `json:"rtype"`
	UploadId   string   `json:"uploadid"`
	BlockList  []string `json:"block_list"`
	LocalCTime int64    `json:"local_ctime"`
	LocalMTime int64    `json:"local_mtime"`
	ZipQuality int      `json:"zip_quality"`
	ZipSign    string   `json:"zip_sign"`
	IsRevision int      `json:"is_revision"`
	Mode       int      `json:"mode"`
	ExifInfo   string   `json:"exif_info"`
}

func (p *CreateFileParam) Values() url.Values {
	values := url.Values{}
	values.Add("path", p.Path)
	values.Add("size", fmt.Sprintf("%d", p.Size))
	if p.IsDir {
		values.Add("isdir", "1")
	} else {
		values.Add("isdir", "0")
	}

	values.Add("rtype", fmt.Sprintf("%d", p.RType))
	values.Add("uploadid", p.UploadId)
	if p.BlockList != nil && len(p.BlockList) > 0 {
		values.Add("block_list", fmt.Sprintf("[\"%s\"]", strings.Join(p.BlockList, "\",\"")))
	}
	if p.LocalCTime > 0 {
		values.Add("local_ctime", fmt.Sprintf("%d", p.LocalCTime))
	}
	if p.LocalMTime > 0 {
		values.Add("local_mtime", fmt.Sprintf("%d", p.LocalMTime))
	}
	if p.ZipQuality > 0 {
		values.Add("zip_quality", fmt.Sprintf("%d", p.ZipQuality))
	}
	if p.ZipSign != "" {
		values.Add("zip_sign", p.ZipSign)
	}
	if p.Mode > 0 {
		values.Add("mode", fmt.Sprintf("%d", p.Mode))
	}
	if p.ExifInfo != "" {
		values.Add("exif_info", p.ExifInfo)
	}
	return values
}

func (p *CreateFileParam) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

type CreateFile struct {
	UploadFileInfo
	ErrNo          int    `json:"errno"`
	ServerFilename string `json:"server_filename"`
}

func (p *CreateFile) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}
