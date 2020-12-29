package xmlrpc

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"
)

const (
	contentType = `text/xml`
)

// MethodCall ...
type MethodCall struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     []Param  `xml:"params>param"`
}

// MethodResponse ...
type MethodResponse struct {
	XMLName xml.Name `xml:"methodResponse"`
	Params  *[]Param `xml:"params>param,omitempty"`
	Fault   *Value   `xml:"fault>value,omitempty"`
}

// Param ...
type Param struct {
	Value Value
}

// Value ...
type Value struct {
	XMLName xml.Name  `xml:"value"`
	Int     *int      `xml:"int,omitempty"`
	String  *string   `xml:"string,omitempty"`
	Members *[]Member `xml:"struct>member,omitempty"`
}

// Member ...
type Member struct {
	XMLName xml.Name `xml:"member"`
	Name    string   `xml:"name"`
	Value   Value
}

// FaultError ...
func FaultError(v *Value) error {
	if v == nil {
		return nil
	}
	if v.Members != nil && len(*v.Members) == 2 {
		m0, m1 := (*v.Members)[0], (*v.Members)[1]
		if m0.Name == `faultCode` && m1.Name == `faultString` {
			if m0.Value.Int != nil && m1.Value.String != nil {
				return fmt.Errorf(
					`xmlrpc: error: faultCode=%d, faultString=%s`,
					*m0.Value.Int, *m1.Value.String,
				)
			}
		}
	}
	return fmt.Errorf(`xmlrpc: unknown fault error`)
}

// Send ...
func Send(ctx context.Context, server string, payload *MethodCall) (*MethodResponse, error) {
	u, err := url.Parse(server)
	if err != nil {
		zap.L().Info(`xmlrpc: bad server url`, zap.Error(err), zap.String(`url`, server))
		return nil, err
	}
	x, err := xml.MarshalIndent(payload, ``, `  `)
	if err != nil {
		zap.L().Info(`xmlrpc: error marshaling payload`, zap.Error(err))
		return nil, err
	}

	contentLength := len(x)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, // An XML-RPC message is an HTTP-POST request.
		u.String(), bytes.NewReader(x),
	)
	if err != nil {
		zap.L().Info(`xmlrpc: error creating request`, zap.Error(err))
		return nil, err
	}

	// The Content-Type is text/xml.
	req.Header.Set(`Content-Type`, contentType)

	// The Content-Length must be specified and must be correct.
	req.Header.Set(`Content-Length`, strconv.Itoa(contentLength))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.L().Info(`xmlrpc: request failed`, zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Unless there's a lower-level error, always return 200 OK.
	if resp.StatusCode != 200 {
		zap.L().Info(`xmlrpc: request failed`, zap.String("status", resp.Status))
		return nil, fmt.Errorf(`xmlrpc: error status: %s`, resp.Status)
	}

	// The Content-Type is text/xml.
	ct := resp.Header.Get("Content-Type")
	if !isContentTypeXML(ct) {
		zap.L().Info(`xmlrpc: invalid content type`, zap.String("content-type", ct))
		return nil, fmt.Errorf(`xmlrpc: invalid content type: %s`, ct)
	}

	// Content-Length must be present and correct.
	// Ignore.

	val := MethodResponse{}
	if err = xml.NewDecoder(resp.Body).Decode(&val); err != nil {
		zap.L().Info(`xmlrpc: error unmarshaling response`, zap.Error(err))
		return nil, err
	}

	return &val, nil
}

func isContentTypeXML(ct string) bool {
	mt, _, _ := mime.ParseMediaType(ct) // mt is lower-cased.
	return mt == contentType
}

// NewIntValue ...
func NewIntValue(i int) Value { return Value{Int: &i} }

// NewStringValue ...
func NewStringValue(s string) Value { return Value{String: &s} }

// ResponseWriter ...
type ResponseWriter interface {
	WriteString(msg string)
	WriteFault(code int, msg string)
}

type _ResponseWriter struct {
	wrote bool
	w     http.ResponseWriter
}

func (r *_ResponseWriter) WriteString(msg string) {
	resp := MethodResponse{
		Params: &[]Param{
			{
				Value: NewStringValue(msg),
			},
		},
	}
	r.write(&resp)
}

func (r *_ResponseWriter) WriteFault(code int, msg string) {
	faultResp := MethodResponse{
		Fault: &Value{
			Members: &[]Member{
				{
					Name:  `faultCode`,
					Value: NewIntValue(code),
				},
				{
					Name:  `faultString`,
					Value: NewStringValue(msg),
				},
			},
		},
	}
	r.write(&faultResp)
}

func (r *_ResponseWriter) write(v interface{}) {
	if r.wrote {
		return
	}
	b, _ := xml.MarshalIndent(v, ``, `  `)
	r.w.Header().Set(`Content-Type`, contentType)
	r.w.Header().Set(`Content-Length`, strconv.Itoa(len(b)))
	r.w.Write(b)
	r.wrote = true
}

// Handler ...
func Handler(fn func(w ResponseWriter, method string, args []Param)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			zap.L().Info(`xmlrpc: invalid method`, zap.String("method", r.Method))
			http.Error(w, `XML-RPC server accepts POST requests only.`, http.StatusMethodNotAllowed)
			return
		}

		rw := &_ResponseWriter{w: w}

		// The Content-Type is text/xml.
		ct := r.Header.Get("Content-Type")
		if !isContentTypeXML(ct) {
			zap.L().Info(`xmlrpc: invalid content type`, zap.String("Content-Type", ct))
			rw.WriteFault(-1, `invalid content type`)
			return
		}

		call := MethodCall{}
		if err := xml.NewDecoder(r.Body).Decode(&call); err != nil {
			zap.L().Info(`xmlrpc: malformed request`, zap.Error(err))
			rw.WriteFault(-1, `malformed request`)
			return
		}

		if call.MethodName == `` {
			zap.L().Info(`xmlrpc: method name required`)
			rw.WriteFault(-1, `method name required`)
			return
		}

		fn(rw, call.MethodName, call.Params)
	})
}
