package libvalid

import "unsafe"

func IsNil(parameter interface{}) bool {
	// see https://dev.to/pauljlucas/go-tcha-when-nil--nil-hic
	return parameter == nil || (*[2]uintptr)(unsafe.Pointer(&parameter))[1] == 0
}

func NotNil(parameter interface{}) bool {
	return !IsNil(parameter)
}

func AnyNil(values ...interface{}) bool {
	for _, val := range values {
		if IsNil(val) {
			return true
		}
	}
	return false
}
