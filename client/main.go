package main

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v2"
)

func initHostConfigs() HostConfig {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(usr.HomeDir, "/.taoblog.yml")
	fp, err := os.Open(path)
	if err != nil {
		panic("cannot read init config: " + path)
	}
	defer fp.Close()

	hostConfigs := map[string]HostConfig{}
	ymlDec := yaml.NewDecoder(fp)
	if err := ymlDec.Decode(&hostConfigs); err != nil {
		panic(err)
	}

	// select which host to use
	host := os.Getenv("HOST")
	if host == "" {
		host = "blog"
	}
	hostConfig, ok := hostConfigs[host]
	if !ok {
		panic("cannot find init config for host: " + host)
	}
	return hostConfig
}

func main() {
	flag.Parse()

	config := initHostConfigs()
	client := NewClient(config)

	if len(os.Args) >= 2 {
		command := os.Args[1]
		switch command {
		case "get", "post", "delete", "patch":
			client.CRUD(command, os.Args[2])
		case "posts":
			switch os.Args[2] {
			case "init":
				client.InitPost()
			case "create":
				client.CreatePost()
			case "upload":
				client.UploadPostFiles(os.Args[3:])
			case "update":
				client.UpdatePost()
			case "pub", "publish":
				client.SetPostStatus("public")
			case "draft":
				client.SetPostStatus("draft")
			case "get":
				client.GetPost()
			default:
				panic("unknown operation")
			}
		case "comments":
			switch os.Args[2] {
			default:
				panic("unknown operation")
			case "set-post-id":
				cmtID, err := strconv.ParseInt(os.Args[3], 10, 0)
				if err != nil {
					panic(err)
				}
				postID, err := strconv.ParseInt(os.Args[4], 10, 0)
				if err != nil {
					panic(err)
				}
				client.SetCommentPostID(cmtID, postID)
			}
		case "backup":
			client.Backup(os.Stdout)
		case "settings":
			client.Settings(os.Args[2:])
		default:
			panic("unknown operation")
		}
	}
}
