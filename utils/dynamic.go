package utils

import "reflect"

// func Call(instance interface{}, method string, params []interface{}) []reflect.Value {
// 	in := make([]reflect.Value, len(params))
// 	for k, param := range params {
// 		in[k] = reflect.ValueOf(param)
// 	}
// 	return reflect.ValueOf(instance).MethodByName(method).Call(in)
// }

func Call(instance interface{}, method string, params []reflect.Value) []reflect.Value {
	return reflect.ValueOf(instance).MethodByName(method).Call(params)
}
