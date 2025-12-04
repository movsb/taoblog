package micros_utils

import (
	"log"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/modules/renderers/auto_image_border"
)

func (u *Utils) SetAutoImageBorderCreator(fn func() *auto_image_border.Task) {
	u.autoImageBorderCreator = fn
}

func (u *Utils) RegisterAutoImageBorderHandler(s proto.Utils_RegisterAutoImageBorderHandlerServer) (outErr error) {
	defer utils.CatchAsError(&outErr)

	user.MustBeAdmin(s.Context())

	log.Println(`接入图片边框处理`)

	task := u.autoImageBorderCreator()

	task.Run(s.Context(), s)

	return nil
}
