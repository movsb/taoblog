package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/movsb/taoblog/client"
	"github.com/movsb/taoblog/server"
	"github.com/spf13/cobra"
)

var (
	builtOn   string
	builtAt   string
	goVersion string
	gitAuthor string
	gitCommit string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: `TaoBlog client & server program.`,
	}

	versionCmd := &cobra.Command{
		Use:   `version`,
		Short: `Show version`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stderr, ""+
				"Built On  : %s\n"+
				"Built At  : %s\n"+
				"Go Version: %s\n"+
				"Git Author: %s\n"+
				"Git Commit: %s\n",
				builtOn, builtAt,
				goVersion, gitAuthor, gitCommit,
			)
		},
	}

	rootCmd.AddCommand(versionCmd)

	client.AddCommands(rootCmd)
	server.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
