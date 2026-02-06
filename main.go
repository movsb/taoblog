package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/cmd/imports"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/cmd/tools"
	"github.com/movsb/taoblog/cmd/tools/conv"
	"github.com/movsb/taoblog/modules/version"
	_ "github.com/movsb/taoblog/setup/tool-deps"
	"github.com/spf13/cobra"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	version.ForceEnableDevMode = os.Getenv(`DEV`)
	cobra.EnableCommandSorting = false
}

func main() {
	rootCmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: `TaoBlog client & server program.`,
	}

	server.AddCommands(rootCmd)
	client.AddCommands(rootCmd)
	imports.AddCommands(rootCmd)
	tools.AddCommands(rootCmd)
	conv.AddCommands(rootCmd)
	version.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
