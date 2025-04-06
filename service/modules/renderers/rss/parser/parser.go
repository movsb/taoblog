package rss_parser

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type Trimmed string

func (t Trimmed) String() string {
	return strings.TrimSpace(string(t))
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel struct {
		XMLName       xml.Name `xml:"channel"`
		Title         Trimmed  `xml:"title"`
		Link          Trimmed  `xml:"link"`
		Description   Trimmed  `xml:"description"`
		LastBuildDate Date     `xml:"lastBuildDate"`
		Items         []struct {
			XMLName     xml.Name `xml:"item"`
			Title       Trimmed  `xml:"title"`
			Link        Trimmed  `xml:"link"`
			PubDate     Date     `xml:"pubDate"`
			Description Trimmed  `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

type Date struct {
	time.Time
}

func (t *Date) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	layouts := [...]string{time.RFC1123, time.RFC1123Z, time.RFC3339}
	for _, layout := range layouts {
		tt, err := time.Parse(layout, s)
		if err == nil {
			t.Time = tt
			return nil
		}
	}

	return fmt.Errorf(`cannot parse time: %v`, s)
}

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Head    struct {
		Title Trimmed `xml:"title"`
	} `xml:"head"`
	Body struct {
		Outlines []Outline `xml:"outline"`
	} `xml:"body"`
}

type Outline struct {
	Title    Trimmed   `xml:"title,attr"`
	Type     Trimmed   `xml:"type,attr"`
	XmlUrl   Trimmed   `xml:"xmlUrl,attr"`
	Outlines []Outline `xml:"outline"`
}

func (opml *OPML) Each(fn func(title, url string)) {
	m := map[string]struct{}{}
	var pr func(outline []Outline)
	pr = func(outlines []Outline) {
		for _, outline := range outlines {
			if outline.Type == `rss` {
				if _, ok := m[outline.XmlUrl.String()]; !ok {
					m[outline.XmlUrl.String()] = struct{}{}
					fn(outline.Title.String(), outline.XmlUrl.String())
				}
			}
			pr(outline.Outlines)
		}
	}
	pr(opml.Body.Outlines)
}

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Title   Trimmed  `xml:"title"`
	Updated Date     `xml:"updated"`
	Entries []struct {
		Title Trimmed `xml:"title"`
		Link  struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Published Date `xml:"published"`
		Summary   struct {
			Type string  `xml:"type,attr"`
			Data Trimmed `xml:",chardata"`
		} `xml:"summary"`
		Content struct {
			Type string  `xml:"type,attr"`
			Data Trimmed `xml:",chardata"`
		} `xml:"content"`
	} `xml:"entry"`
}
