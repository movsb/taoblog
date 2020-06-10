package version

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// BuiltOn ...
	BuiltOn string
	// BuiltAt ...
	BuiltAt string
	// GoVersion ...
	GoVersion string
	// GitAuthor ...
	GitAuthor string
	// GitCommit ...
	GitCommit string
)

// AddCommands ...
func AddCommands(rootCmd *cobra.Command) {
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
				BuiltOn, BuiltAt,
				GoVersion, GitAuthor, GitCommit,
			)
		},
	}

	rootCmd.AddCommand(versionCmd)
}
