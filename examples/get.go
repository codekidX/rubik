package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rubikorg/rubik"
)

func timerHook(rc *rubik.Context) {
	start := time.Now()
	go func() {
		<-rc.AfterChan
		var d int64
		var metric string
		if time.Since(start).Milliseconds() <= 0 {
			d = time.Since(start).Microseconds()
			metric = "us"
		} else {
			d = time.Since(start).Milliseconds()
			metric = "ms"
		}
		fmt.Printf("Hook: %d%s | %s\n", d, metric, "/")
	}()
}

func testfunc(c *rubik.Context) {
	c.JSON(http.StatusOK, rubik.RouteTree{})
}

func main() {
	rubik.GET("/", timerHook, testfunc).
		Doc(`root api for the app`).
		Name("index route").
		Group("root")

	// rubik.Hook(timerHook)
	err := rubik.Run()
	if err != nil {
		panic(err)
	}

}
