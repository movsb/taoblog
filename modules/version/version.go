package version

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	Name          = `TaoBlog`
	NameLowercase = `taoblog`
)

// 运行起始时间。
var Time = time.Now()

// 在运行时控制开发者模式。
// 只对部分可修改的配置生效，方便用于测试。
var ForceEnableDevMode string

// 是否是开发模式
func DevMode() bool {
	if ForceEnableDevMode != `` {
		return slices.Contains([]string{`1`, `yes`, `true`}, ForceEnableDevMode)
	}

	exists := func(f string) bool {
		_, err := os.Stat(f)
		return err == nil
	}

	var envDev bool
	switch {
	case GitCommit == ``:
		envDev = true
	case strings.EqualFold(GitCommit, `head`):
		envDev = true
	case exists(`go.mod`):
		envDev = true
	}

	return envDev
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
				"Git Commit: %s\n"+
				"\n"+
				"Dev Mode  : %v"+
				"\n",
				BuiltOn, BuiltAt,
				GoVersion, GitAuthor, GitCommit,
				DevMode(),
			)
		},
	}

	rootCmd.AddCommand(versionCmd)
}
