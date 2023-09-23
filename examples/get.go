package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rubikorg/rubik"
)

func timer(rc *rubik.Context) {
	start := time.Now()
	go func() {
		<-rc.AfterChan
		var d int64
		var metric string
		if time.Now().Sub(start).Milliseconds() <= 0 {
			d = time.Now().Sub(start).Microseconds()
			metric = "ns"
		} else {
			d = time.Now().Sub(start).Milliseconds()
			metric = "ms"
		}
		fmt.Printf("%d%s | %s\n", d, metric, rc.Request.URL)
	}()
}

func main() {
	rubik.Use(rubik.Route{
		Path:   "/",
		Method: []string{http.MethodGet},
		Responders: []rubik.Responder{
			timer,
			func(c *rubik.Context) { c.JSON(http.StatusOK, rubik.RouteTree{}) },
		},
	})
	rubik.Run()
}
