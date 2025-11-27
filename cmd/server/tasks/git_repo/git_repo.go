package git_repo

import (
	"context"
	"log"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	backups_git "github.com/movsb/taoblog/modules/backups/git"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
)

// client：带凭证的客户端。
func Sync(ctx context.Context, client *clients.ProtoClient, notifier proto.NotifyServer, setLastSyncedAt func(time.Time)) {
	gs := backups_git.New(ctx, client, false)

	sync := func() error {
		if version.DevMode() {
			log.Println(`开发模式不运行 git 同步`)
			return nil
		}

		err := gs.Sync()
		if err == nil {
			log.Println(`git 同步成功`)
			setLastSyncedAt(time.Now())
		} else {
			notifier.SendInstant(
				auth.SystemForLocal(context.Background()),
				&proto.SendInstantRequest{
					Title: `同步失败`,
					Body:  err.Error(),
					Group: `同步与备份`,
					Level: proto.SendInstantRequest_Active,
				},
			)
		}

		return err
	}

	const every = time.Hour * 1
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println(`git 同步任务退出`)
			return
		case <-gs.Do():
			log.Println(`立即执行同步中`)
			if err := sync(); err != nil {
				log.Println(err)
			}
		case <-ticker.C:
			if err := sync(); err != nil {
				log.Println(err)
			}
		}
	}
}
