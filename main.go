package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/movsb/taoblog/client"
	"github.com/movsb/taoblog/server"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: `TaoBlog client & server program.`,
	}

	client.AddCommands(rootCmd)
	server.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
