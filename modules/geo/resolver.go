package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	urlpkg "net/url"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
)

type GeoLocationResolver interface {
	ResolveGeoLocation(ctx context.Context, latitude, longitude float32) ([]string, error)
}

type Baidu struct {
	ak       string
	referrer func() string
}

func NewBaidu(ak string, referrer func() string) *Baidu {
	return &Baidu{ak, referrer}
}

type PointOfInterest struct {
	Name     string `json:"name"`
	Distance int    `json:"distance,string"`
}
type BaiduResponse struct {
	Status int `json:"status"`
	Result struct {
		PointOfInterests []PointOfInterest `json:"pois"`
	} `json:"result"`
}

// curl 'https://api.map.baidu.com/geocoder/v2/?ak=${AK}&output=json&location=${LATITUDE},${LNGITUDE}&coordtype=wgs84ll&pois=1' -H 'Referer: https://blog.twofei.com'
func (b *Baidu) ResolveGeoLocation(ctx context.Context, latitude, longitude float32) (_ []string, outErr error) {
	defer utils.CatchAsError(&outErr)

	url := utils.Must1(urlpkg.Parse(`https://api.map.baidu.com/geocoder/v2/?ak=&output=json&location=&coordtype=wgs84ll&pois=1`))
	args := url.Query()
	args.Set(`ak`, b.ak)
	args.Set(`location`, fmt.Sprintf(`%f,%f`, latitude, longitude))
	url.RawQuery = args.Encode()

	req := utils.Must1(http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil))
	req.Header.Add(`Referer`, b.referrer())
	req.Header.Add(`User-Agent`, version.Name)

	rsp := utils.Must1(http.DefaultClient.Do(req))
	if rsp.StatusCode != 200 {
		panic(fmt.Sprintf(`failed to resolve: %s`, rsp.Status))
	}
	defer rsp.Body.Close()

	var br BaiduResponse
	utils.Must(json.NewDecoder(rsp.Body).Decode(&br))

	if br.Status != 0 {
		panic(fmt.Sprintf(`geo status error: %d`, br.Status))
	}

	pois := br.Result.PointOfInterests
	slices.SortFunc(pois, func(a, b PointOfInterest) int {
		return a.Distance - b.Distance
	})

	return utils.Map(pois, func(p PointOfInterest) string {
		return p.Name
	}), nil
}
