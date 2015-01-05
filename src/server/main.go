package main

import (
	"app/register"
	"fmt"
	"gslib"
	"gslib/utils/store"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1)

	store.InitSharedInstance()

	start := time.Now()
	times := 10000000
	for i := 0; i < times; i++ {
		// ets.Set("aaaa", "hello world")
		// ets.Get("aaaa")
		store.Del("aaaa")
	}
	duration := time.Since(start)
	fmt.Println("used time: ", duration.Seconds(), " Per Second: ", float64(times)/duration.Seconds())

	register.Load()
	gslib.Run()
}
