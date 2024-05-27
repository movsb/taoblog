package twitter

import (
	"log"
	"os"

	"github.com/movsb/taoblog/protocols/clients"
	"github.com/spf13/cobra"
)

func CreateCommands(client func() *clients.ProtoClient) *cobra.Command {
	twitterCmd := &cobra.Command{
		Use:   `twitter`,
		Short: `twitter <dir>`,
		Run: func(cmd *cobra.Command, args []string) {
			root, err := cmd.Flags().GetString(`dir`)
			if err != nil {
				log.Fatalln(err)
			}
			if root == "" {
				log.Fatalln(`没指定数据根目录。`)
			}
			withoutAssets, err := cmd.Flags().GetBool(`no-assets`)
			if err != nil {
				log.Fatalln(err)
			}
			importer := New(os.DirFS(root), client())
			if err := importer.Execute(withoutAssets); err != nil {
				log.Fatalln(err)
			}
		},
	}
	twitterCmd.Flags().Bool(`no-assets`, false, `不上传附件。`)
	twitterCmd.Flags().StringP(`dir`, `d`, ``, `推特导出数据的根目录`)
	return twitterCmd
}
