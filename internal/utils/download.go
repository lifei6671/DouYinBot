package utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

var DefaultUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0`
var ErrAnimatedWebP = errors.New("animated webp")

// DownloadCover 下载封面到指定路径
func DownloadCover(authorId, urlStr, filename string) (string, error) {
	uri, err := url.ParseRequestURI(urlStr)
	if err != nil {
		logs.Error("解析封面文件失败: url[%s] filename[%s] %+v", urlStr, filename, err)
		return "", err
	}

	hash := md5.Sum([]byte(uri.Path))
	hashStr := hex.EncodeToString(hash[:])

	ext := filepath.Ext(uri.Path)

	filename = filepath.Join(filename, authorId, "cover", hashStr+ext)

	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}
	f, err := os.Create(filename)
	if err != nil {
		logs.Error("创建封面文件失败: url[%s] filename[%s] %+v", urlStr, filename, err)
		return "", err
	}
	defer SafeClose(f)

	header := http.Header{}
	header.Add("Accept", "*/*")
	header.Add("Accept-Encoding", "identity;q=1, *;q=0")
	header.Add("User-Agent", DefaultUserAgent)
	header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6,mt;q=0.5,ru;q=0.4,de;q=0.3")
	header.Add("Referer", urlStr)
	header.Add("Pragma", "no-cache")

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		logs.Error("下载封面文件失败: url[%s] filename[%s] %+v", urlStr, filename, err)
		return "", err
	}
	req.Header = header
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}
	defer SafeClose(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("http error: status_code[%d] err_msg[%s]", resp.StatusCode, string(b))
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		logs.Error("保存图片失败: %s  %+v", urlStr, err)
		return "", err
	}
	if ext == "" {
		switch resp.Header.Get("Content-Type") {
		case "image/jpeg":
			ext = ".jpeg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
		newPath := filename + ext
		if ext == ".webp" {
			if ok, err := IsAnimatedWebP(filename); ok && err == nil {
				_ = os.Remove(filename)
				return "", ErrAnimatedWebP
			}
		}

		if err := os.Rename(filename, newPath); err == nil {

			filename = newPath
		}

	}

	if ext != ".webp" {
		newPath := strings.TrimSuffix(filename, ext) + ".webp"
		if oErr := Image2Webp(filename, newPath); oErr == nil {
			_ = os.Remove(filename)
			return newPath, nil
		} else {
			logs.Error("转换 WebP 格式出错： %+v", oErr)
		}
	}

	logs.Info("保存封面成功: %s  %s", urlStr, filename)
	return filename, nil
}
