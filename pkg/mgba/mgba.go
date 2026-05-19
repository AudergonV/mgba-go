// go-mgba: Go bindings for mGBA
// Copyright (c) 2026 Vincent Audergon, MPL-2.0 License
// This project provides Go bindings for the mGBA emulator (https://mgba.io/).

package mgba

/*
#cgo LDFLAGS: -lmgba
#cgo CFLAGS: -I/usr/include/mgba
#include "mgba.h"
*/
import "C"

func Init() {
	C.mgba_init()
}
