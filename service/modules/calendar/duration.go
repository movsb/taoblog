package calendar

import (
	"fmt"
	"regexp"
	"strconv"
)

type Unit byte

const (
	UnitDay = iota
	UnitWeek
	UnitMonth
	UnitYear
)

type Duration struct {
	N    int
	Unit Unit
}

var reSplitDuration = regexp.MustCompile(`^([1-9][0-9]*)([d|w|m|y])$`)

func ParseDuration(d string) (Duration, error) {
	matches := reSplitDuration.FindStringSubmatch(d)
	if matches == nil {
		return Duration{}, fmt.Errorf(`无法解析为周期：%s`, d)
	}

	var dd Duration
	dd.N, _ = strconv.Atoi(matches[1])

	switch matches[2] {
	case `d`:
		dd.Unit = UnitDay
	case `w`:
		dd.Unit = UnitWeek
	case `m`:
		dd.Unit = UnitMonth
	case `y`:
		dd.Unit = UnitYear
	}

	return dd, nil
}
