package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var sourceNames = []string{
	"README.md",
	"index.md",
	"index.html",
}

func update(id int64, args map[string]string) {

	var err error
	buf := bytes.NewBuffer(nil)
	mpw := multipart.NewWriter(buf)

	// arg: pid
	err = mpw.WriteField("pid", fmt.Sprint(id))
	if err != nil {
		panic(err)
	}

	// args to be posted
	for k, a := range args {
		err = mpw.WriteField(k, a)
		if err != nil {
			panic(err)
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !initConfig.verify,
			},
		},
	}

	mpw.Close()

	req, err := http.NewRequest("POST", initConfig.api+"/posts/update", buf)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", mpw.FormDataContentType())
	req.Header.Set("Authorization", initConfig.key)
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

func readSource(dir string) (string, string) {
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

func readPostMetas(dir string) map[string]string {
	path := filepath.Join(dir, "metas")
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	metas := make(map[string]string)
	buf := bufio.NewScanner(fp)
	for buf.Scan() {
		line := strings.TrimSpace(buf.Text())
		if line == "" {
			continue
		}
		toks := strings.SplitN(line, ":", 2)
		if len(toks) < 2 {
			log.Printf("invalid meta: %s\n", line)
			continue
		}
		metas[toks[0]] = toks[1]
	}
	return metas
}

func evalPost(args []string) {
	if len(args) >= 1 {
		if args[0] == "update" {
			if len(args) >= 2 {
				dir := args[1]
				metas := readPostMetas(dir)
				pid, err := strconv.ParseInt(metas["id"], 10, 64)
				if err != nil {
					panic(err)
				}
				delete(metas, "id")
				typ, src := readSource(dir)
				metas["source_type"] = typ
				metas["source"] = src
				update(pid, metas)
			}
		}
	}
}
