package pingback

import (
	"context"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	buf, err := ioutil.ReadAll(io.LimitReader(resp.Body, bufSize))
	if err != nil {
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

// Handler ...
func Handler(fn func(w xmlrpc.ResponseWriter, source, target string, title string)) http.HandlerFunc {
	return xmlrpc.Handler(func(w xmlrpc.ResponseWriter, r *xmlrpc.Request) {
		if r.MethodName != pingbackMethod {
			zap.L().Info(`pingback: unknown method`, zap.String(`method`, r.MethodName))
			w.WriteFault(0, `unknown method`)
			return
		}
		if len(r.Args) != 2 {
			zap.L().Info(`pingback: two args required`)
			w.WriteFault(0, `two args required`)
			return
		}

		var source, target *string
		if p := r.Args[0].Value.String; p != nil {
			source = p
		}
		if p := r.Args[1].Value.String; p != nil {
			target = p
		}
		if source == nil || target == nil {
			zap.L().Info(`pingback: two string args required`)
			w.WriteFault(0, `both source and target must be of type string`)
			return
		}

		if !checkURLs(w, r, *source, *target) {
			return
		}

		ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		var title string
		if err := verifySource(ctxTimeout, *source, *target, &title); err != nil {
			zap.L().Info(`pingback: no reference found`)
			w.WriteFault(0, `no reference found`)
			return
		}

		fn(w, *source, *target, title)
	})
}

func checkURLs(w xmlrpc.ResponseWriter, r *xmlrpc.Request, source string, target string) bool {
	sourceURL, err := url.Parse(source)
	if err != nil {
		zap.L().Info(`pingback: invalid source`, zap.String(`source`, source))
		w.WriteFault(0, `invalid source`)
		return false
	}
	switch sourceURL.Scheme {
	case `http`, `https`:
		break
	default:
		zap.L().Info(`pingback: invalid source scheme`, zap.String(`source`, source))
		w.WriteFault(0, `invalid source scheme`)
		return false
	}

	remoteIP, _, _ := net.SplitHostPort(r.Req.RemoteAddr)
	ips, _ := net.LookupHost(sourceURL.Hostname())
	ipFound := false
	for _, ip := range ips {
		if ip == remoteIP {
			ipFound = true
			break
		}
	}
	if !ipFound {
		zap.L().Info(`pingback: invalid source url`,
			zap.String(`source`, source),
			zap.String(`remote_addr`, r.Req.RemoteAddr),
			zap.Strings(`ips`, ips),
		)
		w.WriteFault(0, `invalid source url`)
		return false
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		zap.L().Info(`pingback: invalid target`, zap.String(`target`, target))
		w.WriteFault(0, `invalid target`)
		return false
	}
	switch targetURL.Scheme {
	case `http`, `https`:
		break
	default:
		zap.L().Info(`pingback: invalid target scheme`, zap.String(`target`, target))
		w.WriteFault(0, `invalid target scheme`)
		return false
	}

	if strings.EqualFold(sourceURL.Host, targetURL.Host) {
		zap.L().Info(`pingback: source host == target host, ignored.`, zap.String(`target`, target))
		w.WriteFault(0, `source host == target host`)
		return false
	}

	return true
}

func verifySource(ctx context.Context, source string, target string, title *string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		zap.L().Info(`pingback: verifySource: failed`, zap.Error(err))
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.L().Info(`pingback: verifySource: failed`, zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	const maxBodySize = 1 << 20
	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		zap.L().Info(`pingback: verifySource: parse html failed`, zap.Error(err))
		return err
	}

	doc.Find(`title`).Each(func(i int, s *goquery.Selection) {
		*title = strings.TrimSpace(s.Text())
	})
	if *title == `` {
		zap.L().Info(`pingback: verifySource: empty title`, zap.String(`source`, source))
		return fmt.Errorf(`empty title`)
	}

	found := false

	doc.Find(`a`).Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr(`href`)
		// simple compare
		if href == target {
			found = true
			return
		}
	})

	if !found {
		zap.L().Info(`pingback: verifySource: no reference found`, zap.String(`source`, source))
		return fmt.Errorf(`no reference found`)
	}

	return nil
}
