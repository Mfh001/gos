package main

import (
	"app/register"
	"gslib"
)

type hello struct {
	name string
	age  int32
}

func main() {
	register.Load()
	gslib.Run()
}
