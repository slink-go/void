package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

func getQueryParam(ctx *gin.Context, key string) string {
	v, ok := ctx.Request.URL.Query()[key]
	if !ok {
		return ""
	}
	return v[0]
}
func queryParams(ctx *gin.Context, joiner string) string {
	var result []string
	params := ctx.Request.URL.Query()
	if len(params) > 0 {
		for k, p := range params {
			for _, v := range p {
				result = append(result, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}
	if len(result) == 0 {
		return ""
	}
	return fmt.Sprintf("?%s", strings.Join(result, joiner))
}
