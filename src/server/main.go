package main

import (
	"app/register"
	"gslib"
	"gslib/store"
)

func main() {
	register.Load()
	store.InitDB()
	register.DataLoaderMap["equips"]("54B9D9C02B897851E30F71F4", nil)
	gslib.Run()
}
