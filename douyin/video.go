package douyin

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Video struct {
	PlayId      string
	PlayAddr    string
	Cover       string
	OriginCover string
	MusicAddr   string
	Author      struct {
		Id           string
		Nickname     string
		AvatarLarger string
	}
}

func (v *Video) GetFilename() string {
	return v.Author.Nickname + "-" + v.PlayId + ".mp4"
}

func (v *Video) Download(filename string) error {
	name := v.Author.Nickname + "-" + v.PlayId + ".mp4"
	f, err := os.Stat(filename)
	if err == nil && f.IsDir() {
		filename = filepath.Join(filename, name)
	} else if os.IsNotExist(err) {
		var dir string
		if filepath.Ext(filename) == "" {
			dir = filename
			filename = filepath.Join(filename, name)
		} else {
			dir = filepath.Dir(filename)
		}
		if err := os.MkdirAll(dir, 0655); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodGet, v.PlayAddr, nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f1, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f1.Close()
	_, err = io.Copy(f1, resp.Body)
	return err
}

func (v *Video) GetDownloadUrl() (string, error) {
	req, err := http.NewRequest(http.MethodGet, v.PlayAddr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	lv := resp.Header.Get("Location")

	return lv, nil
}
