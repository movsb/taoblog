package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	urlpkg "net/url"

	wgs2gcj "github.com/googollee/eviltransform/go"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
)

type GeoLocationResolver interface {
	ResolveGeoLocation(ctx context.Context, latitude, longitude float32) ([]string, error)
}

type GaoDe struct {
	ak string
}

func NewGeoDe(ak string) *GaoDe {
	return &GaoDe{ak}
}

type Location struct {
	Name     string  `json:"name"`
	Distance float32 `json:"distance,string"`
}

type Response struct {
	Status    int `json:"status,string"`
	RegeoCode struct {
		AOIs []Location `json:"aois"`
		POIs []Location `json:"pois"`
	} `json:"regeocode"`
}

// https://lbs.amap.com/api/webservice/guide/api/georegeo#t5
func (g *GaoDe) ResolveGeoLocation(ctx context.Context, latitude, longitude float32) (_ []string, outErr error) {
	defer utils.CatchAsError(&outErr)

	gcjLatitude, gcjLongitude := wgs2gcj.WGStoGCJ(float64(latitude), float64(longitude))

	url := utils.Must1(urlpkg.Parse(`https://restapi.amap.com/v3/geocode/regeo?key=&location=&radius=250&extensions=all&output=json`))
	args := url.Query()
	args.Set(`key`, g.ak)
	args.Set(`location`, fmt.Sprintf(`%f,%f`, gcjLongitude, gcjLatitude))
	url.RawQuery = args.Encode()
	req := utils.Must1(http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil))
	req.Header.Add(`User-Agent`, version.Name)
	rsp := utils.Must1(http.DefaultClient.Do(req))
	if rsp.StatusCode != 200 {
		panic(fmt.Sprintf(`failed to resolve: %s`, rsp.Status))
	}
	defer rsp.Body.Close()

	var geoRsp Response
	utils.Must(json.NewDecoder(rsp.Body).Decode(&geoRsp))

	if geoRsp.Status != 1 {
		panic(fmt.Sprintf(`geo status error: %d`, geoRsp.Status))
	}

	locations := []Location{}
	locations = append(locations, geoRsp.RegeoCode.AOIs...)
	locations = append(locations, geoRsp.RegeoCode.POIs...)

	// 去排序去重。
	slices.SortFunc(locations, func(a, b Location) int {
		return strings.Compare(a.Name, b.Name)
	})
	locations = slices.CompactFunc(locations, func(a, b Location) bool {
		return a.Name == b.Name
	})
	// 然后把距离排序。
	slices.SortFunc(locations, func(a, b Location) int {
		return int(a.Distance - b.Distance)
	})

	return utils.Map(locations, func(loc Location) string {
		return loc.Name
	}), nil
}
