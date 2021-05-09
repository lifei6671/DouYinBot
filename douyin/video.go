package douyin

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Video struct {
	PlayId       string
	PlayAddr     string
	VideoRawAddr string
	PlayRawAddr  string
	Cover        string
	OriginCover  string
	MusicAddr    string
	Desc         string
	Author       struct {
		Id           string
		ShortId      string
		Nickname     string
		AvatarLarger string
		Signature    string
	}
}

func (v *Video) GetFilename() string {
	return FilterEmoji(v.Author.Nickname) + "-" + v.PlayId + ".mp4"
}

func (v *Video) Download(filename string) (string, error) {
	name := filepath.Join(v.Author.Id, v.GetFilename())
	f, err := os.Stat(filename)
	if err == nil && f.IsDir() {
		filename = filepath.Join(filename, name)
	}
	dir := filepath.Dir(filename)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0655); err != nil {
			return "", err
		}
	}

	req, err := http.NewRequest(http.MethodGet, v.PlayAddr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	f1, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f1.Close()
	_, err = io.Copy(f1, resp.Body)
	return filename, err
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
