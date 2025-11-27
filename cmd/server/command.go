package server

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/service/modules/request_throttler"
	"github.com/spf13/cobra"
)

func AddCommands(rootCmd *cobra.Command) {
	var (
		monitorDomainInitialDelay bool
	)

	serveCommand := &cobra.Command{
		Use:   `server`,
		Short: `Run the server`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var cfg *config.Config
			dir := os.DirFS(`.`)
			demo := utils.Must1(cmd.Flags().GetBool(`demo`))
			if demo {
				cfg = config.DefaultDemoConfig()
				// 并且强制关闭本地环境。
				version.ForceEnableDevMode = `0`
			} else {
				cfg2 := config.DefaultConfig()
				if err := config.ApplyFromFile(cfg2, dir, `taoblog.yml`); err != nil {
					if !os.IsNotExist(err) && !errors.Is(err, io.EOF) {
						log.Fatalln(err)
					}
				}
				cfg = cfg2
			}
			configOverride := func(cfg *config.Config) {
				if err := config.ApplyFromFile(cfg, dir, `taoblog.override.yml`); err != nil {
					if !os.IsNotExist(err) {
						log.Fatalln(err)
					}
				}
			}

			s := NewServer(
				WithRequestThrottler(request_throttler.New()),
				WithCreateFirstPost(),
				WithGitSyncTask(true),
				WithBackupTasks(true),
				WithRSS(true),
				WithMonitorCerts(true),
				WithMonitorDomain(true, monitorDomainInitialDelay),
				WithConfigOverride(configOverride),
				WithYearProgress(),
			)

			s.Serve(context.Background(), false, cfg, nil)
		},
	}

	serveCommand.Flags().Bool(`demo`, false, `运行演示实例。`)
	serveCommand.Flags().BoolVar(&monitorDomainInitialDelay, `test-monitor-domain-initial-delay`, true, `是否启用首次域名检测延时等待。`)

	serveCommand.Flags().SortFlags = false
	rootCmd.AddCommand(serveCommand)
}
