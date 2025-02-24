package imports

import (
	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/cmd/imports/files"
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
		return clients.NewProtoClient(config.Home, config.Token)
	}))

	importsCmd.AddCommand(files.CreateCommands())

	parent.AddCommand(importsCmd)
}
