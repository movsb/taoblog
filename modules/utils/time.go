package utils

import "time"

type CurrentTimezoneGetter interface {
	GetCurrentTimezone() *time.Location
}

type LocalTimezoneGetter struct{}

func (LocalTimezoneGetter) GetCurrentTimezone() *time.Location {
	return time.Local
}
