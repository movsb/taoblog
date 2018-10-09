package main

import (
	"fmt"
	"net/http"
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

// EndReq succs or fails a request.
func EndReq(c *gin.Context, err interface{}, dat interface{}) {
	succ := false

	if err == nil {
		succ = true
	} else if val, ok := err.(error); ok {
		succ = val == nil
	} else if val, ok := err.(bool); ok {
		succ = val
	} else if val, ok := err.(string); ok {
		succ = val == ""
	}

	if succ {
		c.JSON(http.StatusOK, dat)
	} else {
		if e, ok := err.(error); ok {
			c.JSON(http.StatusInternalServerError, e.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err,
				"dat": dat,
			})
		}
	}
}
