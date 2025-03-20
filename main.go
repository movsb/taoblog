package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/cmd/daemon"
	"github.com/movsb/taoblog/cmd/imports"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	_ "github.com/movsb/taoblog/setup/tool-deps"
	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rootCmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: `TaoBlog client & server program.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			version.ForceEnableDevMode = utils.Must1(cmd.Flags().GetString(`dev`))
		},
	}
	// 但是仍然会判断 git 提交号。
	rootCmd.PersistentFlags().String(`dev`, ``, `强制设置调试模式开关。`)

	version.AddCommands(rootCmd)
	client.AddCommands(rootCmd)
	server.AddCommands(rootCmd)
	daemon.AddCommands(rootCmd)
	imports.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
