package main

import (
	"app/register"
	"fmt"
	"gslib"
	"gslib/utils"
)

func main() {
	register.Load()
	register.RegisterDataLoader()
	register.CustomRegisterDataLoader()
	gslib.Run()
}
