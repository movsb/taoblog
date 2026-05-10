package utils_test

import (
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

func TestRelativeDate(t *testing.T) {
	now := time.Date(2026, 5, 10, 16, 25, 30, 0, time.Local)
	for _, tc := range []struct {
		t string
		r string
	}{
		{t: `2026-05-10 16:25:05`, r: `刚刚`},
		{t: `2026-05-10 16:24:05`, r: `1分钟前`},
		{t: `2026-05-10 16:23:05`, r: `2分钟前`},
		{t: `2026-05-10 15:25:05`, r: `1小时前`},
		{t: `2026-05-10 00:00:00`, r: `16小时前`},
		{t: `2026-05-09 23:59:59`, r: `昨天`},
		{t: `2026-05-09 00:00:00`, r: `昨天`},
		{t: `2026-05-08 23:59:59`, r: `前天`},
		{t: `2026-05-08 00:00:00`, r: `前天`},
		{t: `2026-05-07 23:59:59`, r: `3天前`},
		{t: `2026-05-07 00:00:00`, r: `3天前`},
		{t: `2026-04-10 00:00:00`, r: `30天前`},
		{t: `2026-04-09 00:00:00`, r: `04月09日`},
		{t: `2025-04-09 00:00:00`, r: `2025年04月09日`},
	} {
		tt := utils.Must1(time.ParseInLocation(time.DateTime, tc.t, time.Local))
		if r := utils.RelativeDateFrom(tt, now); r != tc.r {
			t.Errorf("RelativeDateFrom(%s) = %s, want %s", tc.t, r, tc.r)
		}
	}
}
