package unsafe

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"unsafe"
)

func SameInterfacePointer(handler1, handler2 application.Handler) bool {
	return (*(*[2]uintptr)(unsafe.Pointer(&handler1)))[1] == (*(*[2]uintptr)(unsafe.Pointer(&handler2)))[1]
}
