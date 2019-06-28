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
package packet

import (
	"reflect"
)

//----------------------------------------------- export struct fields with packet writer.
func Pack(tos int16, tbl interface{}, writer *Packet) []byte {
	if writer == nil {
		writer = Writer()
	}

	// write code
	if tos != -1 {
		writer.WriteUint16(uint16(tos))
	}

	// is nil?
	v := reflect.ValueOf(tbl)
	if !v.IsValid() {
		return writer.Data()
	}

	// deal with pointers
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		v = v.Elem()
	}
	count := v.NumField()

	for i := 0; i < count; i++ {
		f := v.Field(i)
		switch f.Type().Kind() {
		case reflect.Slice, reflect.Array:
			writer.WriteUint16(uint16(f.Len()))
			for a := 0; a < f.Len(); a++ {
				if _is_primitive(f.Index(a)) {
					_write_primitive(f.Index(a), writer)
				} else {
					elem := f.Index(a).Interface()
					Pack(-1, elem, writer)
				}
			}
		case reflect.Struct:
			Pack(-1, f.Interface(), writer)
		default:
			_write_primitive(f, writer)
		}
	}

	return writer.Data()
}

//----------------------------------------------- test whether the field is primitive type
func _is_primitive(f reflect.Value) bool {
	switch f.Type().Kind() {
	case reflect.Bool,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	}
	return false
}

//----------------------------------------------- write a primitive field
func _write_primitive(f reflect.Value, writer *Packet) {
	switch f.Type().Kind() {
	case reflect.Bool:
		writer.WriteBool(f.Interface().(bool))
	case reflect.Uint8:
		writer.WriteByte(f.Interface().(byte))
	case reflect.Uint16:
		writer.WriteUint16(f.Interface().(uint16))
	case reflect.Uint32:
		writer.WriteUint32(f.Interface().(uint32))
	case reflect.Uint64:
		writer.WriteUint64(f.Interface().(uint64))

	case reflect.Int:
		writer.WriteUint32(uint32(f.Interface().(int)))
	case reflect.Int8:
		writer.WriteByte(byte(f.Interface().(int8)))
	case reflect.Int16:
		writer.WriteUint16(uint16(f.Interface().(int16)))
	case reflect.Int32:
		writer.WriteUint32(uint32(f.Interface().(int32)))
	case reflect.Int64:
		writer.WriteUint64(uint64(f.Interface().(int64)))

	case reflect.Float32:
		writer.WriteFloat32(f.Interface().(float32))

	case reflect.Float64:
		writer.WriteFloat64(f.Interface().(float64))

	case reflect.String:
		writer.WriteString(f.Interface().(string))
	}
}
