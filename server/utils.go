package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
			if e == sql.ErrNoRows {
				c.Status(404)
				return
			}
			c.JSON(http.StatusInternalServerError, e.Error())
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err,
				"dat": dat,
			})
		}
	}
}
