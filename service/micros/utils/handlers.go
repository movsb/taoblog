package micros_utils

import (
	"io"
	"log"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (u *Utils) RegisterAutoImageBorderHandler(s proto.Utils_RegisterAutoImageBorderHandlerServer) (outErr error) {
	defer utils.CatchAsError(&outErr)

	user.MustBeAdmin(s.Context())

	if u.autoImageBorderCreator == nil {
		return status.Error(codes.FailedPrecondition, `服务未正确初始化。`)
	}
	if u.currentAutoImageBorderHandler.Load() {
		return status.Error(codes.AlreadyExists, `已经有处理器正在处理。`)
	}

	log.Println(`接入图片边框处理`)
	defer log.Println(`退出图片边框处理`)

	u.currentAutoImageBorderHandler.Store(true)
	defer u.currentAutoImageBorderHandler.Store(false)

	task := u.autoImageBorderCreator()
	task.Run(s.Context(), func(file *models.File, input io.Reader, r, g, b byte, ratio float32) (_ float32, outErr error) {
		defer utils.CatchAsError(&outErr)

		log.Println(`发送数据`)

		utils.Must(s.Send(&proto.AutoImageBorderRequest{
			PostId: uint32(file.PostID),
			Path:   file.Path,

			Data: utils.Must1(io.ReadAll(input)),

			R: uint32(r),
			G: uint32(g),
			B: uint32(b),

			Ratio: ratio,
		}))

		log.Println(`等待接收数据`)

		return utils.Must1(s.Recv()).GetValue(), nil
	})

	return nil
}
