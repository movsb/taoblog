package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

var sourceNames = []string{
	"README.md",
	"index.md",
	"index.html",
}

var key string
var dir string
var api string
var verify bool

func update(id int64, typ string, source string) {

	var err error
	buf := bytes.NewBuffer(nil)
	mpw := multipart.NewWriter(buf)

	// arg: pid
	err = mpw.WriteField("pid", fmt.Sprint(id))
	if err != nil {
		panic(err)
	}

	// arg: type
	err = mpw.WriteField("type", typ)
	if err != nil {
		panic(err)
	}

	// arg: source
	err = mpw.WriteField("source", source)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !verify,
			},
		},
	}

	mpw.Close()

	req, err := http.NewRequest("POST", api+"/posts/update-content", buf)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", mpw.FormDataContentType())
	req.Header.Set("Authorization", key)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	code := resp.StatusCode
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("code: %d\nbody: %s\n", code, body)
}

func readSource() (string, string) {
	var source string
	var theName string

	for _, name := range sourceNames {
		path := filepath.Join(dir, name)
		bys, err := ioutil.ReadFile(path)
		if err != nil {
			continue
		}
		source = string(bys)
		theName = name
		break
	}
	if source == "" {
		panic("source cannot be found")
	}

	typ := ""
	switch filepath.Ext(theName) {
	case ".md":
		typ = "markdown"
	case ".html":
		typ = "html"
	}

	return typ, source
}

func readID() int64 {
	path := filepath.Join(dir, "id")
	bys, err := ioutil.ReadFile(path)
	if err != nil || len(bys) == 0 {
		panic(err)
	}
	str := strings.SplitN(string(bys), "\n", 2)[0]
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		panic(err)
	}
	return id
}

func parseFlags() {
	flag.StringVar(&key, "key", "", "api key")
	flag.StringVar(&dir, "dir", ".", "post dir")
	flag.StringVar(&api, "api", "", "api")
	flag.BoolVar(&verify, "verify", true, "verify host key")
	flag.Parse()
}

func main() {
	parseFlags()

	id := readID()
	typ, src := readSource()
	update(id, typ, src)
}
