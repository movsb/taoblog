package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
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

	req, err := http.NewRequest("GET", initConfig.api+"/backups/backup", nil)
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

	dec := json.NewDecoder(resp.Body)
	ret := JSONRet{}
	err = dec.Decode(&ret)
	if err != nil {
		panic(err)
	}
	fmt.Print(ret.Data)
}

func evalBackup(args []string) {
	if len(args) >= 1 {
		if args[0] == "backup" {
			doBackup()
		}
	}
}
