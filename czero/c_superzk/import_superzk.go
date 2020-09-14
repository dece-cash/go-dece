package c_superzk

import _ "github.com/dece-cash/go-dece/czero/lib"

/*

#cgo CFLAGS: -I ../lib/szk_include

#cgo LDFLAGS: -L ../lib/lib -lsuperzk

*/
import "C"

func Is_czero_debug() bool {
	return false
}
