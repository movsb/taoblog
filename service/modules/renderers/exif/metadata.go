package exif

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// exiftool -G -s -json test_data/exif.avif
type Metadata struct {
	MimeType     string  `json:"File:MIMEType"`         // 类型：image/avif
	FileName     string  `json:"File:FileName"`         // 文件名字
	FileSize     string  `json:"File:FileSize"`         // 文件大小
	ImageSize    string  `json:"Composite:ImageSize"`   // 尺寸
	Model        string  `json:"EXIF:Model"`            // 设置型号
	Make         string  `json:"EXIF:Make"`             // 设置制造商
	FNumber      float32 `json:"EXIF:FNumber"`          // 光圈数
	FocalLength  string  `json:"EXIF:FocalLength"`      // 焦距
	ExposureTime string  `json:"EXIF:ExposureTime"`     // 曝光时间
	GPSPosition  string  `json:"Composite:GPSPosition"` // 坐标
	GPSAltitude  string  `json:"Composite:GPSAltitude"` // 海拔
	CreateDate   string  `json:"EXIF:CreateDate"`       // 创建日期/时间
	OffsetTime   string  `json:"EXIF:OffsetTime"`       // 时区
}

func (m *Metadata) CreationDateTime() time.Time {
	if m.CreateDate == "" {
		return time.Time{}
	}
	timeZone := time.Now()
	if m.OffsetTime != "" {
		t, err := time.Parse(`-07:00`, m.OffsetTime)
		if err != nil {
			log.Println(`failed to parse timezone for exif:`, m.OffsetTime)
			return time.Time{}
		}
		timeZone = t
	}
	layout := `2006:01:02 15:04:05`
	t, err := time.ParseInLocation(layout, m.CreateDate, timeZone.Location())
	if err != nil {
		log.Println(`failed to parse time:`, m.CreateDate)
		return time.Time{}
	}
	return t
}

func (m *Metadata) String() []string {
	var pairs []string

	add := func(value string, name string) {
		if value != "" {
			pairs = append(pairs, name, value)
		}
	}

	add(m.FileName, `名字`)
	add(m.FileSize, `大小`)
	add(m.ImageSize, `尺寸`)
	add(m.MimeType, `类型`)

	if t := m.CreationDateTime(); !t.IsZero() {
		f := t.Format(time.RFC3339)
		add(f, `时间`)
	}

	if m.Make != "" && m.Model != "" {
		add(m.Make+` / `+m.Model, `设备`)
	}

	lenInfo := []string{}
	if m.FNumber > 0 {
		lenInfo = append(lenInfo, fmt.Sprintf(`f/%v`, m.FNumber))
	}
	if m.FocalLength != "" {
		lenInfo = append(lenInfo, m.FocalLength)
	}
	if m.ExposureTime != "" {
		lenInfo = append(lenInfo, m.ExposureTime)
	}
	add(strings.Join(lenInfo, `, `), `镜头`)

	add(m.GPSPosition, `位置`)
	add(m.GPSAltitude, `海拔`)

	return pairs
}
