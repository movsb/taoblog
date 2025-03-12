package imports

import (
	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/cmd/imports/twitter"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/spf13/cobra"
)

func AddCommands(parent *cobra.Command) {
	importsCmd := &cobra.Command{
		Use: `imports`,
	}

	importsCmd.AddCommand(twitter.CreateCommands(func() *clients.ProtoClient {
		config := client.InitHostConfigs()
		return clients.NewProtoClientFromHome(config.Home, config.Token)
	}))

	parent.AddCommand(importsCmd)
}
