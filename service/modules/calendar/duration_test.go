package calendar_test

import (
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/calendar"
)

func TestParseDuration(t *testing.T) {
	tcs := []struct {
		d string
		p calendar.Duration
	}{
		{`1d`, calendar.Duration{1, calendar.UnitDay}},
		{`2w`, calendar.Duration{2, calendar.UnitWeek}},
		{`3m`, calendar.Duration{3, calendar.UnitMonth}},
		{`4y`, calendar.Duration{4, calendar.UnitYear}},
	}
	for _, tc := range tcs {
		dd := utils.Must1(calendar.ParseDuration(tc.d))
		if dd != tc.p {
			t.Error(`not equal:`, tc.d, tc.p)
		}
	}
}
