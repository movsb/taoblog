package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/movsb/taoblog/client"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/server"
	_ "github.com/movsb/taoblog/setup/tool-deps"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	rootCmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: `TaoBlog client & server program.`,
	}

	version.AddCommands(rootCmd)
	client.AddCommands(rootCmd)
	server.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
