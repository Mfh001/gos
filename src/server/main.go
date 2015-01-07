package main

import (
	"app/register"
	"fmt"
	"gslib"
	"gslib/store"
	"reflect"
	"runtime"
	"time"
)

type hello struct {
	name string
	age  int32
}

func main() {
	runtime.GOMAXPROCS(8)

	store.InitSharedInstance()
	store.Test()

	start := time.Now()
	times := 10
	person := hello{
		name: "savin",
		age:  26,
	}

	rp := reflect.ValueOf(person)
	fmt.Println("reflect: ", rp.Type().Field(1).Name)

	// for i := 0; i < times; i++ {
	// 	// ets.Set("aaaa", "hello world")
	// 	// ets.Get("aaaa")
	// 	store.Set([]string{"persons"}, "person", person)
	// 	store.Get([]string{"persons"}, "person")
	// }
	// duration := time.Since(start)
	// fmt.Println("used time: ", duration.Seconds(), " Per Second: ", float64(times)/duration.Seconds())

	store.Set([]string{"p", "persons"}, "person", person)
	store.Set([]string{"p", "persons"}, "person1", person)
	store.Set([]string{"p", "persons"}, "person2", person)
	vperson1 := store.Get([]string{"p", "persons"}, "person2").(hello)
	vperson1.name = "chaned savin"

	vperson := store.Get([]string{"p", "persons"}, "person2").(hello)
	fmt.Println("vperson: ", vperson.name)

	for i := 0; i < times; i++ {
		store.Select([]string{"p", "persons"}, func(elem interface{}) bool {
			v := elem.(hello)
			return v.name == "savin"
		})
	}
	duration := time.Since(start)
	fmt.Println("used time: ", duration.Seconds(), " Per Second: ", float64(times)/duration.Seconds())

	// for k, v := range persons {
	// 	if k == 1 {
	// 		store.Del([]string{"p", "persons"}, "person")
	// 	}
	// 	fmt.Println("name: ", v.(*hello).name)
	// }

	// setFunc := func() {
	// 	for {
	// 		vperson := store.Get("player1", "person").(*hello)
	// 		fmt.Println("addr: ", vperson.name)
	// 	}
	// }

	// go setFunc()

	// for i := 0; i < 1000000000; i++ {
	// 	person.name = fmt.Sprintf("changed to: %d", i)
	// }

	// fmt.Println("----------------------", vperson.name)
	// store.Set("player1", "person", person)
	// person.name = "changed savin"
	// vperson := store.Get("player1", "person").(*hello)
	// fmt.Println("addr: ", vperson.name)
	// fmt.Println("----------------------")

	register.Load()
	gslib.Run()
}
