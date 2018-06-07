package main

import (
	"time"
)

var datetime DateTime

const (
	mysqlFormat   = "2006-01-02 15:04:05"
	httpgmtFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
)

type DateTime struct {
}

// MyGMT is the UTC time
func (DateTime) MyGMT() string {
	return time.Now().UTC().Format(mysqlFormat)
}

// MyLocal is the Local time
func (DateTime) MyLocal() string {
	return time.Now().Format(mysqlFormat)
}

func (DateTime) My2HTTPGMT(t string) string {
	tim, err := time.Parse(mysqlFormat, t)
	if err != nil {
		panic(err)
	}
	return tim.Format(httpgmtFormat)
}

func (DateTime) HTTPGMT2My(t string) string {
	tim, err := time.Parse(httpgmtFormat, t)
	if err != nil {
		panic(err)
	}
	return tim.Format(mysqlFormat)
}

func (DateTime) My2Local(t string) string {
	tim, err := time.Parse(mysqlFormat, t)
	if err != nil {
		panic(err)
	}
}
