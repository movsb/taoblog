package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"

	"github.com/golang/protobuf/jsonpb"
	"github.com/movsb/taoblog/protocols"
	"nhooyr.io/websocket"
)

type FileSystemWrapper struct {
	marshaler   *jsonpb.Marshaler
	unmarshaler *jsonpb.Unmarshaler
}

func NewFileSystemWrapper() *FileSystemWrapper {
	return &FileSystemWrapper{
		marshaler: &jsonpb.Marshaler{
			OrigName:     true,
			EmitDefaults: true,
		},
		unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: false,
		},
		// 这个对 oneof 解析有误，暂时不用。
		// marshaller: &runtime.JSONPb{
		// 	MarshalOptions: protojson.MarshalOptions{
		// 		UseProtoNames:   true,
		// 		EmitUnpopulated: true,
		// 	},
		// },
	}
}

// TODO 好像这是一个通用的 GRPC Stream <--> WebSocket 包装方法？🤔
func (fs *FileSystemWrapper) fileServer(ctx context.Context, ws *websocket.Conn, fsc protocols.Management_FileSystemClient) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer fsc.CloseSend()
		for {
			ty, r, err := ws.Reader(ctx)
			if err != nil {
				log.Println(err)
				return
			}
			if ty != websocket.MessageText {
				log.Println(`invalid message type`)
				return
			}
			req := protocols.FileSystemRequest{}
			if err := fs.unmarshaler.UnmarshalNext(json.NewDecoder(r), &req); err != nil {
				log.Println(err)
				return
			}
			if _, err := r.Read(nil); !errors.Is(err, io.EOF) {
				log.Println(`extra message`)
				return
			}
			if err := fsc.Send(&req); err != nil {
				log.Println(err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			rsp, err := fsc.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				log.Println(err)
				return
			}
			w, err := ws.Writer(ctx, websocket.MessageText)
			if err != nil {
				log.Println(err)
				return
			}
			if err := fs.marshaler.Marshal(w, rsp); err != nil {
				log.Println(err)
				return
			}
			if err := w.Close(); err != nil {
				log.Println(err)
				return
			}
		}
	}()

	wg.Wait()
}
