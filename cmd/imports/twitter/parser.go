package twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
)

func ParseTweets(fsys fs.FS) ([]*Tweet, error) {
	all, err := read(fsys, `data/tweets.js`)
	if err != nil {
		return nil, err
	}

	var w []struct {
		Tweet *Tweet `json:"tweet"`
	}

	if err := json.Unmarshal(all, &w); err != nil {
		return nil, err
	}

	tweets := make([]*Tweet, 0, len(w))
	for _, w := range w {
		tweets = append(tweets, w.Tweet)
	}

	for _, t := range tweets {
		if t.ID == `1653346313336127489` {
			t.ID += ""
		}
		findChildren(tweets, t)
	}

	return tweets, nil
}

func findChildren(all []*Tweet, t *Tweet) {
	replies := []*Tweet{}
	for _, r := range all {
		if r.InReplyToStatusID != "" && r.InReplyToStatusID == t.ID {
			replies = append(replies, r)
		}
	}
	t.children = replies
	for _, r := range replies {
		findChildren(all, r)
	}
}

func read(fsys fs.FS, path string) ([]byte, error) {
	fp, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	all, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}
	stripped, err := strip(all)
	if err != nil {
		return nil, err
	}
	return stripped, nil
}

func strip(raw []byte) ([]byte, error) {
	eq := bytes.IndexByte(raw, '=')
	br := bytes.IndexByte(raw, '[')
	nl := bytes.IndexByte(raw, '\n')
	if eq >= 0 && br >= 0 && (br < nl || nl == -1) {
		return raw[br:], nil
	}
	return nil, fmt.Errorf(`格式不正确。`)
}
