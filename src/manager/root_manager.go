package manager

import (
	"fmt"
	"runtime"
)

type Params struct {
	Path string
	Name string
	Age  int
}

type RootManager struct {
}

// gen_server callbacks
func (self *RootManager) Init(name string) (err error) {
	fmt.Println("server ", name, " started!")
	return nil
}

//gen_server callbacks
func(self *RootManager) HandleCast(args []interface{}) {
  fmt.Println("HandleCast callback")
}

// gen_server callbacks
func (self *RootManager) Terminate(reason string) (err error) {
	fmt.Println("callback Termiante!")
	return nil
}

/*
	IPC Methods
*/

func (self *RootManager) SystemInfo(from string, time int) int {
	// fmt.Println("CPU: ", runtime.NumCPU())
	return runtime.NumCPU()
}

func (self *RootManager) Echo(words string) string {
	// a,b:=0,1;
	// for i := 1; i < 100000; i++ {
	//   a++
	//   b++
	// }
	return words
}
