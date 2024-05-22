package imports

import (
	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/cmd/imports/twitter"
	proto "github.com/movsb/taoblog/protocols"
	"github.com/spf13/cobra"
)

func AddCommands(parent *cobra.Command) {
	importsCmd := &cobra.Command{
		Use: `imports`,
	}

	importsCmd.AddCommand(twitter.CreateCommands(func() *proto.ProtoClient {
		config := client.InitHostConfigs()
		return proto.NewProtoClient(
			proto.NewConn(config.API, config.GRPC),
			config.Token,
		)
	}))

	parent.AddCommand(importsCmd)
}
