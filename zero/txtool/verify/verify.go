package verify

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txs/stx"
	"github.com/dece-cash/go-dece/zero/txs/zstate"
	"github.com/dece-cash/go-dece/zero/txtool/verify/verify_1"
)

func VerifyWithoutState(ehash *c_type.Uint256, tx *stx.T, num uint64) (e error) {
	return verify_1.VerifyWithoutState(ehash, tx, num)
}

func VerifyWithState(tx *stx.T, state *zstate.ZState, num uint64) (e error) {
	return verify_1.VerifyWithState(tx, state)
}
