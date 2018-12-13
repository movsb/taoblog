package main

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
)

func doBackup() {
	var err error

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !initConfig.verify,
			},
		},
	}

	req, err := http.NewRequest("GET", initConfig.api+"/backups", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", initConfig.key)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic("bad code:" + resp.Status)
	}
	io.Copy(os.Stdout, resp.Body)
}

func evalBackup(args []string) {
	if len(args) >= 1 {
		if args[0] == "backup" {
			doBackup()
		}
	}
}
