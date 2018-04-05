package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func joinInts(ints []int64, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ints)), delim), "[]")
}

func strInSlice(str []string, s string) bool {
	for _, b := range str {
		if b == s {
			return true
		}
	}
	return false
}

func finishDone(c *gin.Context, code int, msgs string, data interface{}) {
	c.JSON(200, &xJSONRet{code, msgs, data})
}

func finishError(c *gin.Context, code int, err error) {
	c.JSON(200, &xJSONRet{code, fmt.Sprint(err), nil})
}
