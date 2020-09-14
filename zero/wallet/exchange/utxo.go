package exchange

import (
	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txs/assets"
)

type Utxo struct {
	Pkr    c_type.PKr
	Root   c_type.Uint256
	TxHash c_type.Uint256
	Nil    c_type.Uint256
	Num    uint64
	Asset  assets.Asset
	IsZ    bool
	Ignore bool
	flag   int
}

func (utxo *Utxo) NilTxType() string {
	if c_superzk.IsSzkNil(&utxo.Nil) {
		return "SZK"
	} else {
		return "CZERO"
	}
}
