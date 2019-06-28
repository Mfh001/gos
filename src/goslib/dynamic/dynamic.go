/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package dynamic

import "reflect"

func Call(instance interface{}, method string, params []reflect.Value) []reflect.Value {
	return reflect.ValueOf(instance).MethodByName(method).Call(params)
}

func CallWithArgs(instance interface{}, method string, params ...interface{}) []reflect.Value {
	in := ToReflectValues(params)
	return reflect.ValueOf(instance).MethodByName(method).Call(in)
}

func ToReflectValues(args []interface{}) []reflect.Value {
	in := make([]reflect.Value, len(args))
	for k, arg := range args {
		in[k] = reflect.ValueOf(arg)
	}
	return in
}

