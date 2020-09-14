package exchange

import (
	"math/big"

	"github.com/dece-cash/go-dece/zero/txs/assets"

	"github.com/dece-cash/go-dece/zero/txtool/prepare"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txtool"
)

func (self *Exchange) GenTx(param prepare.PreTxParam) (txParam *txtool.GTxParam, e error) {
	txParam, e = prepare.GenTxParam(&param, self, &prepare.DefaultTxParamState{})
	if e == nil && txParam != nil {
		for _, in := range txParam.Ins {
			self.usedFlag.Store(in.Out.Root, 1)
		}
	}
	return
}

func (self *Exchange) buildTxParam(param *prepare.BeforeTxParam) (txParam *txtool.GTxParam, e error) {

	txParam, e = prepare.BuildTxParam(&prepare.DefaultTxParamState{}, param)

	if e == nil && txParam != nil {
		for _, in := range txParam.Ins {
			self.usedFlag.Store(in.Out.Root, 1)
		}
	}
	return
}

func (self *Exchange) FindRoots(pk *c_type.Uint512, currency string, amount *big.Int) (roots prepare.Utxos, remain big.Int) {
	utxos, r := self.findUtxos(pk, currency, amount)
	for _, utxo := range utxos {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	remain = *r
	return
}

func (self *Exchange) FindRootsByTicket(pk *c_type.Uint512, tickets []assets.Ticket) (roots prepare.Utxos, remain map[c_type.Uint256]c_type.Uint256) {
	utxos, remain := self.findUtxosByTicket(pk, tickets)
	for _, utxo := range utxos {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	return
}

func (self *Exchange) DefaultRefundTo(pk *c_type.Uint512) (ret *c_type.PKr) {
	if value, ok := self.accounts.Load(*pk); ok {
		account := value.(*Account)
		return &account.mainPkr
	}
	return nil
}

func (self *Exchange) GetRoot(root *c_type.Uint256) (utxos *prepare.Utxo) {
	if u, e := self.getUtxo(*root); e != nil {
		return nil
	} else {
		return &prepare.Utxo{u.Root, u.Asset}
	}
}
