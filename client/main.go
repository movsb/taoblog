package main

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"

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
		case "backup":
			client.Backup(os.Stdout)
		case "settings":
			client.Settings(os.Args[2:])
		default:
			panic("unknown operation")
		}
	}
}
