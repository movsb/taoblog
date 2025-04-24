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
	"github.com/movsb/taoblog/cmd/tools/conv"
	"github.com/movsb/taoblog/modules/version"
	_ "github.com/movsb/taoblog/setup/tool-deps"
	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	version.ForceEnableDevMode = os.Getenv(`DEV`)

	rootCmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: `TaoBlog client & server program.`,
	}

	version.AddCommands(rootCmd)
	client.AddCommands(rootCmd)
	server.AddCommands(rootCmd)
	daemon.AddCommands(rootCmd)
	imports.AddCommands(rootCmd)
	conv.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
