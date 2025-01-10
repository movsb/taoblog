package daemon

import (
	"context"
	"encoding/gob"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/dialers"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/spf13/cobra"
	"github.com/xtaci/smux"
)

const (
	updateScript  = `docker compose pull taoblog && docker compose up -d taoblog`
	restartScript = `docker compose restart taoblog`
)

func run(script string) error {
	cmd := exec.Command(`bash`, `-c`, script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func AddCommands(parent *cobra.Command) {
	daemonCmd := &cobra.Command{
		Use:   `daemon`,
		Short: `守护进程（更新镜像等）`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			home := utils.Must1(cmd.Flags().GetString(`home`))
			token := utils.Must1(cmd.Flags().GetString(`token`))
			update(home, token)
		},
	}
	daemonCmd.Flags().String(`home`, `http://127.0.0.1:2564`, `首页地址`)
	daemonCmd.Flags().StringP(`token`, `t`, ``, `预共享密钥`)
	parent.AddCommand(daemonCmd)

	remoteDialerCmd := &cobra.Command{
		Use:   `dialer`,
		Short: ``,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			home := utils.Must1(cmd.Flags().GetString(`home`))
			token := utils.Must1(cmd.Flags().GetString(`token`))
			remoteDialer(home, token)
		},
	}
	remoteDialerCmd.Flags().String(`home`, `http://127.0.0.1:2564`, `首页地址`)
	remoteDialerCmd.Flags().StringP(`token`, `t`, ``, `预共享密钥`)
	parent.AddCommand(remoteDialerCmd)
}

func remoteDialer(home string, token string) {
	client := clients.NewProtoClient(home, token)
	drc, err := client.Utils.DialRemote(client.Context())
	if err != nil {
		panic(err)
	}
	defer drc.CloseSend()
	log.Println(`远程连接成功：`, home)
	conn := dialers.NewStreamAsConn(drc)
	session, err := smux.Server(conn, nil)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := session.AcceptStream()
		if err != nil {
			panic(err)
		}
		go func(conn net.Conn) {
			defer conn.Close()

			var open dialers.DialRemoteRequest
			if err := gob.NewDecoder(conn).Decode(&open); err != nil {
				log.Println(err)
				return
			}
			log.Println(`准备连接：`, open.Addr)
			conn2, err := net.Dial("tcp", open.Addr)
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			if err := gob.NewEncoder(conn).Encode(&dialers.DialRemoteResponse{
				Error: errMsg,
			}); err != nil {
				log.Println(`连接失败：`, open.Addr)
				return
			}

			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				defer wg.Done()
				io.Copy(conn, conn2)
			}()
			go func() {
				defer wg.Done()
				io.Copy(conn2, conn)
			}()
			wg.Wait()
		}(conn)
	}
}

func update(home string, token string) {
	client := clients.NewProtoClient(home, token)
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for range ticker.C {
		info, err := client.Blog.GetInfo(context.Background(), &proto.GetInfoRequest{})
		if err != nil {
			if strings.Contains(err.Error(), `502`) {
				log.Println(`需要重启。`)
				if err := run(restartScript); err != nil {
					log.Println(err)
				}
			}
			log.Println(err)
			continue
		}
		if !info.ScheduledUpdate {
			continue
		}
		log.Println(`需要更新。`)
		if err := run(updateScript); err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 10)
	}
}
