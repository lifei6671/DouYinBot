package models

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

var incrID atomic.Int64
var tagReg = regexp.MustCompile(`#([^#\s]+)`)

type DouYinTag struct {
	Id      int       `orm:"column(id);auto;pk"`
	TagID   string    `orm:"column(tag_id);size(255);index;not null" json:"tag_id"`
	Name    string    `orm:"column(name);size(255);index;not null" json:"name"`
	VideoID string    `orm:"column(video_id);size(255);index;description(视频ID)" json:"video_id"`
	Created time.Time `orm:"auto_now_add;type(datetime);description(创建时间)"`
}

func (d *DouYinTag) TableName() string {
	return "douyin_tag"
}

// TableUnique 多字段唯一键
func (d *DouYinTag) TableUnique() [][]string {
	return [][]string{
		{"tag_id", "name", "video_id"},
	}
}

func NewDouYinTag() *DouYinTag {
	return &DouYinTag{}
}

func (d *DouYinTag) Create(text string, videoId string) error {
	// 使用 FindAllString 提取所有匹配的内容
	matches := tagReg.FindAllString(text, -1)

	if len(matches) == 0 {
		return nil
	}
	o := orm.NewOrm()

	for _, m := range matches {
		tagName := strings.TrimSpace(strings.Trim(m, "#"))
		if strings.Contains(tagName, "#") || utf8.RuneCountInString(tagName) > 10 {
			continue
		}
		var tag DouYinTag
		err := o.QueryTable(d.TableName()).Filter("name", tagName).One(&tag)

		newTag := DouYinTag{
			Name:    tagName,
			VideoID: videoId,
			Created: time.Now(),
		}
		//如果没查到
		if errors.Is(err, orm.ErrNoRows) {
			newTag.TagID = strconv.FormatInt(incrID.Add(1), 10)
		} else if err == nil {
			newTag.TagID = tag.TagID
		}
		if newTag.TagID != "" {
			if _, err := o.Insert(&newTag); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DouYinTag) GetList(pageIndex int, tagID string) (list []*DouYinVideo, tagName string, total int, err error) {
	if tagID == "" {
		return
	}
	o := orm.NewOrm()
	offset := (max(pageIndex, 1) - 1) * PageSize
	query := o.QueryTable(d.TableName()).
		OrderBy("-id").
		Filter("tag_id", tagID)

	count, err := query.Count()
	total = int(count)

	var tagList []*DouYinTag
	_, err = query.Offset(offset).Limit(PageSize).All(&tagList)
	var videoIDs []any
	for _, v := range tagList {
		videoIDs = append(videoIDs, v.VideoID)
		tagName = v.Name
	}

	_, err = o.QueryTable(NewDouYinVideo().TableName()).Filter("video_id__in", videoIDs...).All(&list)
	if err != nil {
		return nil, "", 0, err
	}
	return
}

func (d *DouYinTag) GetByID(tagID string) (*DouYinTag, error) {
	err := orm.NewOrm().QueryTable(d.TableName()).Filter("tag_id", tagID).One(&d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DouYinTag) Insert() (int, error) {
	id, err := orm.NewOrm().Insert(d)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (d *DouYinTag) FormatTagHtml(text string) (string, error) {
	// 使用 FindAllString 提取所有匹配的内容
	matches := tagReg.FindAllString(text, -1)

	var tags []any
	for _, m := range matches {
		tags = append(tags, strings.TrimSpace(strings.Trim(m, "#")))
	}
	if len(tags) == 0 {
		return text, nil
	}
	var list []*DouYinTag
	_, err := orm.NewOrm().QueryTable(d.TableName()).Filter("name__in", tags...).All(&list)
	if err != nil {
		return "", err
	}
	for _, v := range list {
		text = strings.ReplaceAll(text, "#"+v.Name+" ", fmt.Sprintf(`<a href="%s" title="%s">#%s</a> `, web.URLFor("TagController.Index", ":tag_id", v.TagID, ":page", 1), v.Name, v.Name))
	}

	return text, nil
}

func init() {
	// 需要在init中注册定义的model
	orm.RegisterModel(new(DouYinTag))

	incrID.Store(time.Now().UnixNano())
}
