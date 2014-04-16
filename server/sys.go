package main

import (
	"runtime"
	"time"
)

import (
	. "utils"
)

//---------------------------------------------------------- system routine
func SysRoutine() {
	// timer
	gc_timer := make(chan int32, 10)
	gc_timer <- 1

	runtime.GC()
	INFO("Goroutine Count:", runtime.NumGoroutine())
	time.AfterFunc(2*time.Second, SysRoutine)
}
