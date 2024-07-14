package daemon

import (
	"context"
	"encoding/gob"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/dialers"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/spf13/cobra"
	"github.com/xtaci/smux"
)

const updateScript = `docker compose pull taoblog && docker compose up -d taoblog`

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
			update()
		},
	}
	parent.AddCommand(daemonCmd)

	remoteDialerCmd := &cobra.Command{
		Use:   `dialer`,
		Short: ``,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			endpoint := utils.Must1(cmd.Flags().GetString(`endpoint`))
			remoteDialer(endpoint)
		},
	}
	remoteDialerCmd.Flags().StringP(`endpoint`, `e`, `https://blog.twofei.com`, "")
	parent.AddCommand(remoteDialerCmd)
}

func remoteDialer(endpoint string) {
	client := clients.NewProtoClient(clients.NewConn(endpoint, ""), "")
	drc, err := client.Utils.DialRemote(client.Context())
	if err != nil {
		panic(err)
	}
	defer drc.CloseSend()
	log.Println(`远程连接成功：`, endpoint)
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

func update() {
	// TODO 写死了。
	client := clients.NewFromGrpcAddr(`127.0.0.1:2563`)
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for range ticker.C {
		info, err := client.GetInfo(context.Background(), &proto.GetInfoRequest{})
		if err != nil {
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
	}
}
