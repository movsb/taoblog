package client

import (
	"bytes"
	"log"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/renderers/auto_image_border"
)

func (c *Client) handleAutoImageBorder() {
	client := utils.Must1(c.Utils.RegisterAutoImageBorderHandler(c.Context()))
	defer client.CloseSend()

	for {
		req, err := client.Recv()
		if err != nil {
			if strings.Contains(err.Error(), `EOF`) {
				return
			}
			log.Println(err)
			return
		}

		log.Printf(`开始处理 %d %s...`, req.PostId, req.Path)

		value := auto_image_border.BorderContrastRatio(
			bytes.NewReader(req.Data),
			byte(req.R), byte(req.G), byte(req.B),
			float64(req.Ratio),
		)

		utils.Must(client.Send(&proto.AutoImageBorderResponse{
			Value: value,
		}))
	}
}
