package features

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/plantuml"
)

//go:embed FEATURES.md
var featuresMd []byte
var featuresTime = time.Now()

var (
	light []byte
	dark  []byte
	lock  sync.RWMutex
)

func New() http.Handler {
	return http.HandlerFunc(features)
}

func features(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	if len(light) == 0 || len(dark) == 0 {
		if err := prepare(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			lock.Unlock()
			return
		}
	}
	lock.Unlock()

	content := utils.IIF(r.PathValue(`theme`) == `light`, light, dark)
	w.Header().Add(`Content-Type`, `image/svg+xml`)
	http.ServeContent(w, r, `features.svg`, featuresTime, bytes.NewReader(content))
}

func prepare() error {
	reFeaturesPlantUML := regexp.MustCompile("```plantuml((?sU).+)```")
	matches := reFeaturesPlantUML.FindSubmatch(featuresMd)
	if len(matches) != 2 {
		return errors.New(`no features found`)
	}
	compressed, err := plantuml.Compress(matches[1])
	if err != nil {
		return err
	}
	light, err = plantuml.Fetch(context.Background(), `https://www.plantuml.com/plantuml`, `svg`, compressed, false)
	if err != nil {
		return err
	}
	dark, err = plantuml.Fetch(context.Background(), `https://www.plantuml.com/plantuml`, `svg`, compressed, true)
	if err != nil {
		return err
	}
	return nil
}
