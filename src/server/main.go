package main

import (
	"app/register"
	"gslib"
)

func main() {
	register.Load()
	register.RegisterDataLoader()
	register.CustomRegisterDataLoader()
	gslib.Run()
}
