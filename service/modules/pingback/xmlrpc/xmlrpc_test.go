package xmlrpc

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"testing"
)

func Test1(t *testing.T) {
	call := MethodCall{
		MethodName: `callName`,
		Params: []Param{
			{
				Value{
					Int: new(int),
				},
			},
			{
				Value{
					Int: new(int),
				},
			},
		},
	}
	b, _ := xml.MarshalIndent(call, ``, `  `)
	fmt.Println(string(b))

	call2 := MethodCall{}
	err := xml.Unmarshal(b, &call2)
	fmt.Printf("%v, %+v, \n", err, call2)
}

func Test2(t *testing.T) {
	resp := MethodResponse{
		Params: &[]Param{
			{
				Value: Value{
					Int: new(int),
				},
			},
		},
		Fault: &Value{
			Members: &[]Member{
				{
					Name: `faultCode`,
					Value: Value{
						Int: new(int),
					},
				},
				{
					Name: `faultString`,
					Value: Value{
						String: new(string),
					},
				},
			},
		},
	}
	b, _ := xml.MarshalIndent(resp, ``, `  `)
	fmt.Println(string(b))

	b = []byte(`<methodResponse>
	<fault>
	  <value>
		<struct>
		  <member>
			<name>faultCode</name>
			<value><int>-32700</int></value>
		  </member>
		  <member>
			<name>faultString</name>
			<value><string>parse error. not well formed</string></value>
		  </member>
		</struct>
	  </value>
	</fault>
  </methodResponse>
  `)

	resp2 := MethodResponse{}
	err := xml.Unmarshal(b, &resp2)
	fmt.Printf("%v, %+v, \n", err, resp2)
}

func TestSend(t *testing.T) {
	t.SkipNow()
	ch := `https://coolshell.cn/xmlrpc.php`
	call := &MethodCall{
		MethodName: `pingback.ping`,
		Params: []Param{
			{
				Value: Value{},
			},
		},
	}
	resp, err := Send(context.TODO(), ch, call)
	if err != nil {
		t.Fatal(err)
	}
	if er := FaultError(resp.Fault); er != nil {
		t.Fatal(er)
	}
	fmt.Println(resp.Params)
}

func TestServe(t *testing.T) {
	t.SkipNow()
	http.HandleFunc(`/xmlrpc`, Handler(func(method string, args []Param) {
		fmt.Println(method, args)
	}))
	http.ListenAndServe(`:8080`, nil)
}
