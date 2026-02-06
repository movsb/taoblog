package client

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/movsb/taoblog/protocols/go/proto"
)

const (
	updateScript = `docker compose pull taoblog && docker compose up -d taoblog`
)

func run(script string) error {
	cmd := exec.Command(`bash`, `-c`, script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func update(client *Client) {
	go func() {
		for {
			memTotal, memAvail := statMemory()
			// TODO 硬编码写/目录了，实质应该是运行时所在目录。
			fsTotal, fsAvail := statDisk(`/`)
			req := proto.SetHostStatesRequest{
				Memory: &proto.SetHostStatesRequest_Memory{
					Total: memTotal,
					Avail: memAvail,
				},
				Filesystem: &proto.SetHostStatesRequest_Filesystem{
					Total: fsTotal,
					Avail: fsAvail,
				},
			}
			_, err := client.Management.SetHostStates(client.Context(), &req)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Minute)
		}
	}()
	for {
		time.Sleep(time.Second * 15)
		info, err := client.Blog.GetInfo(context.Background(), &proto.GetInfoRequest{})
		if err != nil {
			log.Println(`GetInfo:`, err)
			continue
		}
		if !info.ScheduledUpdate {
			continue
		}
		log.Println(`检测到设置了更新标识，需要更新。`)
		if err := run(updateScript); err != nil {
			log.Println(err)
			continue
		}
	}
}
