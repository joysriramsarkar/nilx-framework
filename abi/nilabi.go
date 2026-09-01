//go:build cgo
// +build cgo

// Package abi implements the C ABI boundary for the NilX Runtime.
// It allows embedding NilX in Android (JNI/NDK), iOS (Swift/Metal), and desktop apps.
package abi

/*
#include "nilabi.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

//export nilx_runtime_create
func nilx_runtime_create() C.NilContext {
	id := CreateRuntime()
	return C.NilContext(unsafe.Pointer(id))
}

//export nilx_runtime_destroy
func nilx_runtime_destroy(ctx C.NilContext) {
	id := uintptr(ctx)
	DestroyRuntime(id)
}

//export nilx_runtime_run
func nilx_runtime_run(ctx C.NilContext, nabc *C.uint8_t, nabc_len C.size_t) C.NilResult {
	id := uintptr(ctx)
	data := C.GoBytes(unsafe.Pointer(nabc), C.int(nabc_len))
	err := ExecuteBytecode(id, data)
	var res C.NilResult
	if err != nil {
		res.ok = C.bool(false)
		res.err.code = C.int32_t(-1)
		res.err.message = C.CString(err.Error())
		return res
	}

	res.ok = C.bool(true)
	return res
}

//export nilx_runtime_dispatch_touch
func nilx_runtime_dispatch_touch(ctx C.NilContext, x C.float, y C.float) C.bool {
	id := uintptr(ctx)
	handled := DispatchTouchEvent(id, float64(x), float64(y))
	return C.bool(handled)
}

//export nilx_runtime_get_ui_json
func nilx_runtime_get_ui_json(ctx C.NilContext) *C.char {
	id := uintptr(ctx)
	jsonStr, err := GetUIRootJSON(id)
	if err != nil {
		return C.CString("{}")
	}
	return C.CString(jsonStr)
}
