package micros_utils

import (
	"time"

	"github.com/movsb/taoblog/modules/geo"
	"github.com/movsb/taoblog/service/modules/renderers/auto_image_border"
)

type Option func(u *Utils)

func WithGaoDe(ak string) Option {
	return func(u *Utils) {
		u.geoLocationResolver = geo.NewGeoDe(ak)
	}
}

func WithTimezone(getTimezone func() *time.Location) Option {
	return func(u *Utils) {
		u.timezone = getTimezone
	}
}

func WithAutoImageBorderCreator(fn func() *auto_image_border.Task) Option {
	return func(u *Utils) {
		u.autoImageBorderCreator = fn
	}
}
