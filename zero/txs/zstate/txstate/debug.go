// copyright 2018 The dece.cash Authors
// This file is part of the go-dece library.
//
// The go-dece library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-dece library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-dece library. If not, see <http://www.gnu.org/licenses/>.

package txstate

import (
	"runtime"
	"runtime/debug"

	"github.com/dece-cash/go-dece/czero/deceparam"

	"github.com/dece-cash/go-dece/log"
)

func Need_debug() bool {
	return false
	if false {
		return true
	} else {
		return deceparam.Is_Dev()
	}
}

func Debug_Weak_panic(msg string, ctx ...interface{}) {
	if Need_debug() {
		log.Debug(">========debug_painc:=======>"+msg, ctx...)
		debug.PrintStack()
		runtime.Breakpoint()
	}
}
