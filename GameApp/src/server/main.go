package main

import (
	"app/register"
	"app/register/callbacks"
	"gslib"
	"gslib/leaderboard"
)

func main() {
	register.Load()
	register.RegisterDataLoader()
	register.CustomRegisterDataLoader()
	callbacks.RegisterBroadcast()
	leaderboard.Start()
	gslib.Run()
}
