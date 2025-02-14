package version

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const Name = `TaoBlog`

// 运行起始时间。
var Time = time.Now()

// 在运行时控制开发者模式。
// 只对部分可修改的配置生效，方便用于测试。
var EnableDevMode = true

// 是否是开发模式
func DevMode() bool {
	return EnableDevMode && (GitCommit == "" || strings.EqualFold(GitCommit, `head`))
}

// 在编译脚本里面被注入进来。
var (
	BuiltOn   string
	BuiltAt   string
	GoVersion string
	GitAuthor string
	GitCommit string
)

func init() {
	if GitCommit == `` {
		GitCommit = `HEAD`
	}
}

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
