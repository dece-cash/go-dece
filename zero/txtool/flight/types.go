package flight

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txtool"
)

type PreTxParam struct {
	Gas      uint64
	GasPrice uint64
	From     c_type.PKr
	Ins      []c_type.Uint256
	Outs     []txtool.GOut
}
