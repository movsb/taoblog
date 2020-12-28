package pingback

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"sync"

	"github.com/movsb/taoblog/service/modules/pingback/xmlrpc"
	"go.uber.org/zap"
)

// Header ...
const Header = `X-Pingback`

const pingbackMethod = `pingback.ping`

var reHeader = regexp.MustCompile(`<link rel="pingback" href="([^"]+)" ?/?>`)

// 2.3. Autodiscovery Algorithm
func findServer(ctx context.Context, targetURI string) (server string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURI, nil)
	if err != nil {
		zap.L().Info(`pingback: invalid target uri`, zap.String("target", targetURI), zap.Error(err))
		return ``, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.L().Info(`pingback: request failed`, zap.String("target", targetURI), zap.Error(err))
		return ``, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		zap.L().Info(`pingback: status code != 200`, zap.String("target", targetURI), zap.String("status", resp.Status))
		return ``, fmt.Errorf(`pingback: status code != 200`)
	}

	// 1. Examine the HTTP headers of the response.
	if h := resp.Header.Get(Header); h != "" {
		zap.L().Info(`pingback: found server from header`, zap.String(`server`, h))
		return h, nil
	}

	// 2. Otherwise, search the entity body for the first match of the following regular expression.
	// <link rel="pingback" href="([^"]+)" ?/?>
	const bufSize = 16 << 10
	buf := make([]byte, bufSize)
	if _, err = resp.Body.Read(buf); err != nil {
		zap.L().Info(`pingback: read body failed`, zap.Error(err))
		return ``, err
	}

	matches := reHeader.FindSubmatchIndex(buf)
	if matches == nil {
		zap.L().Info(`pingback: no match found`, zap.String("target", targetURI))
		return ``, fmt.Errorf(`pingback: no match found`)
	}

	server = string(buf[matches[2]:matches[3]])
	server = html.UnescapeString(server)

	return
}

// Ping ...
func Ping(ctx context.Context, source string, targets ...string) {
	if len(targets) <= 0 {
		zap.L().Info(`pingback: no targets to ping`)
		return
	}

	wg := &sync.WaitGroup{}
	for _, target := range targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			ping(ctx, source, t)
		}(target)
	}
	wg.Wait()
}

func ping(ctx context.Context, source string, target string) error {
	server, err := findServer(ctx, target)
	if err != nil {
		zap.L().Info(`pingback: ping: findServer failed`, zap.Error(err))
		return err
	}

	req := xmlrpc.MethodCall{
		MethodName: pingbackMethod,
		Params: []xmlrpc.Param{
			{Value: xmlrpc.NewStringValue(source)},
			{Value: xmlrpc.NewStringValue(target)},
		},
	}

	resp, err := xmlrpc.Send(ctx, server, &req)
	if err != nil {
		zap.L().Info(`pingback: ping: xmlrpc.send failed`, zap.Error(err))
		return err
	}

	if err := xmlrpc.FaultError(resp.Fault); err != nil {
		zap.L().Info(`pingback: ping: xmlrpc fault`, zap.Error(err))
		return err
	}

	zap.L().Info(`pingback: succeeded`, zap.String(`source`, source), zap.String(`target`, target))
	return nil
}
