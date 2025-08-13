package exif_exports

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// exiftool -G -s -json test_data/exif.avif
// https://blog.twofei.com/1442/
//
// Rotation: v12.76 里面是数值“180”，13.30 里面竟然成了字符串“Rotate 180”。🤬
type Metadata struct {
	FileName     string      `json:"File:FileName"`         // 文件名字
	FileSize     string      `json:"File:FileSize"`         // 文件大小
	ImageSize    string      `json:"Composite:ImageSize"`   // 尺寸
	Orientation  string      `json:"EXIF:Orientation"`      // 方向
	Rotation     IntOrString `json:"QuickTime:Rotation"`    // 旋转角度
	MimeType     string      `json:"File:MIMEType"`         // 类型：image/avif
	Artist       string      `json:"EXIF:Artist"`           // 作者
	Copyright    string      `json:"EXIF:Copyright"`        // 版权
	Model        string      `json:"EXIF:Model"`            // 设置型号
	Make         string      `json:"EXIF:Make"`             // 设置制造商
	FNumber      float32     `json:"EXIF:FNumber"`          // 光圈数
	FocalLength  string      `json:"EXIF:FocalLength"`      // 焦距
	ExposureTime string      `json:"EXIF:ExposureTime"`     // 曝光时间
	ISO          int         `json:"EXIF:ISO"`              // 感光度
	GPSPosition  string      `json:"Composite:GPSPosition"` // 坐标
	GPSAltitude  string      `json:"Composite:GPSAltitude"` // 海拔
	CreateDate   string      `json:"EXIF:CreateDate"`       // 创建日期/时间
	OffsetTime   string      `json:"EXIF:OffsetTime"`       // 时区
	Description  string      `json:"EXIF:ImageDescription"` // 图片描述
}

type IntOrString string

func (s *IntOrString) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, (*string)(s)); err == nil {
		return nil
	}
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*s = IntOrString(fmt.Sprint(n))
		return nil
	}
	return fmt.Errorf(`error unmarshaling`)
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

// DJI 的型号在 XML:DroneModel 里面，但是从 JPG 拷贝到 AVIF 的时候带不过去，
// 那就暂时这样简单映射一下？
var knownDJIModels = map[string]string{
	`FC3582`: `Mini 3 Pro`,
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
	add(m.Orientation, `方向`)
	add(m.MimeType, `类型`)

	if t := m.CreationDateTime(); !t.IsZero() {
		f := t.Format(time.RFC3339)
		add(f, `时间`)
	}

	add(m.Artist, `作者`)
	add(m.Copyright, `版权`)
	add(m.Description, `描述`)

	if mapped, ok := knownDJIModels[m.Model]; ok {
		m.Model = mapped
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
		lenInfo = append(lenInfo, m.ExposureTime+`s`)
	}
	if m.ISO > 0 {
		lenInfo = append(lenInfo, fmt.Sprintf(`ISO/%v`, m.ISO))
	}
	add(strings.Join(lenInfo, `, `), `镜头`)

	add(m.GPSPosition, `位置`)
	add(m.GPSAltitude, `海拔`)

	return pairs
}

// https://blog.twofei.com/1618/
func (m *Metadata) SwapSizes() bool {
	switch m.Orientation {
	case `Rotate 90 CW`, `Rotate 270 CW`:
		return true
	}
	switch m.Rotation {
	case `90`, `270`, `Rotate 90`, `Rotate 270`:
		return true
	}
	return false
}
