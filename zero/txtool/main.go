package txtool

import (
	"math/big"

	"github.com/dece-cash/go-dece/zero/txs/assets"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/czero/deceparam"
	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/core/types"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/txs/zstate"
)

type BlockChain interface {
	IsValid() bool
	GetCurrenHeader() *types.Header
	GetHeader(hash *common.Hash) *types.Header
	CurrentState(hash *common.Hash) *zstate.ZState
	IsContract(address common.Address) (ret bool, e error)
	GetDeceGasLimit(to *common.Address, tfee *assets.Token, gasPrice *big.Int) (gaslimit uint64, e error)
	GetTks() []c_type.Tk
	GetTkAt(tk *c_type.Tk) uint64
	GetBlockByNumber(num uint64) *types.Block
	GetHeaderByNumber(num uint64) *types.Header
	GetDB() decedb.Database
}

type Ref struct {
	Bc BlockChain
}

var Ref_inst Ref

func (self *Ref) SetBC(bc BlockChain) {
	self.Bc = bc
}

func (self *Ref) GetDelayedNum(delay uint64) (ret uint64) {
	ret = GetDelayNumber(
		self.Bc.GetCurrenHeader().Number.Uint64(),
		delay,
	)
	return
}

func (self *Ref) CurrentState() (ret *zstate.ZState) {
	defer func() {
		if p := recover(); p != nil {
			num := self.GetDelayedNum(0)
			block := self.Bc.GetBlockByNumber(num)
			hash := block.Hash()
			ret = self.Bc.CurrentState(&hash)
		}
	}()
	num := self.GetDelayedNum(deceparam.DefaultConfirmedBlock())
	block := self.Bc.GetBlockByNumber(num)
	hash := block.Hash()
	ret = self.Bc.CurrentState(&hash)
	return
}

func GetDelayNumber(current uint64, delay uint64) (num uint64) {
	if current < delay {
		return 0
	} else {
		return current - delay
	}
}
