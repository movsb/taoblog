package main

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	fmt.Println(datetime.MyGMT())
	fmt.Println(datetime.MyLocal())
	fmt.Println(datetime.My2HTTPGMT("2018-06-07 13:00:27"))
	fmt.Println(datetime.HTTPGMT2My("Thu, 07 Jun 2018 13:00:27 GMT"))
}
