package grpc_proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/movsb/http2tcp"
)

const PreSharedKey = `not-used-key`

func New(grpcAddress string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, err := http2tcp.Accept(w, r, PreSharedKey)
		if err != nil {
			log.Println(err)
			return
		}

		defer conn.Close()

		remote, err := net.Dial(`tcp`, grpcAddress)
		if err != nil {
			log.Println(err)
			return
		}

		defer remote.Close()

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			io.Copy(conn, remote)
		}()
		go func() {
			defer wg.Done()
			io.Copy(remote, conn)
		}()

		wg.Wait()
	})
}
