package utils

import (
	"reflect"
	"unsafe"
)

// StrToBytes converts string to []byte without memory allocation, but it makes string mutable through resulting []byte.
// Use if you know what you are doing
func StrToBytes(str string) []byte {
	header := (*reflect.StringHeader)(unsafe.Pointer(&str))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: header.Data,
		Len:  header.Len,
		Cap:  header.Len,
	}))
}

// BytesToStr converts string to []byte without memory allocation, but the resulting string can mutate on changing []byte.
// Use if you know what you are doing
func BytesToStr(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

// Noescape usage can lead to unexpected memory errors. Use carefully
func Noescape[T any](obj *T) *T {
	x := uintptr(unsafe.Pointer(obj))
	return (*T)(unsafe.Pointer(x ^ 0))
}
