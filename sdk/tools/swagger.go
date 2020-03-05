package tools

import (
	"github.com/codekidX/cherry"
	"net/http"
)

func SwaggerUI() cherry.Plugin {
	return cherry.Plugin{
		Pattern: "/api",
		Handler: http.FileServer(http.Dir("./swaggerui")),
	}
}
