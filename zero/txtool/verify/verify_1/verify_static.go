package verify_1

import (
	"fmt"

	"github.com/dece-cash/go-dece/zero/utils"

	"github.com/dece-cash/go-dece/czero/c_superzk"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/czero/deceparam"
	"github.com/dece-cash/go-dece/czero/superzk"
	"github.com/dece-cash/go-dece/zero/txs/stx"
	"github.com/dece-cash/go-dece/zero/txtool/verify/verify_utils"
)

type verifyWithoutStateCtx struct {
	tx              *stx.T
	num             uint64
	hash            c_type.Uint256
	oout_count      int
	oin_count       int
	zout_count      int
	zin_count       int
	cin_proof_proc  *utils.Procs
	cout_proof_proc *utils.Procs
	pkg_proof_proc  *utils.Procs
}

func VerifyWithoutState(ehash *c_type.Uint256, tx *stx.T, num uint64) (e error) {
	if *ehash != tx.Ehash {
		e = verify_utils.ReportError("ehash error", tx)
		return
	}
	ctx := verifyWithoutStateCtx{}
	ctx.num = num
	ctx.tx = tx
	return ctx.verify()
}

func (self *verifyWithoutStateCtx) prepare() {
	self.hash = self.tx.Tx1_Hash()
	self.cin_proof_proc = verify_input_procs_pool.GetProcs()
	self.cout_proof_proc = verify_output_procs_pool.GetProcs()
	self.pkg_proof_proc = verify_pkg_procs_pool.GetProcs()
	return
}

func (self *verifyWithoutStateCtx) clear() {
	verify_input_procs_pool.PutProcs(self.cin_proof_proc)
	verify_output_procs_pool.PutProcs(self.cout_proof_proc)
	verify_pkg_procs_pool.PutProcs(self.pkg_proof_proc)
}

func (self *verifyWithoutStateCtx) verifyFee() (e error) {
	if !verify_utils.CheckUint(&self.tx.Fee.Value) {
		e = verify_utils.ReportError("txs.verify check fee too big", self.tx)
		return
	}
	self.tx.ToFeeCC_Szk()
	self.oout_count++
	return
}

func (self *verifyWithoutStateCtx) verifyFrom() (e error) {
	if !superzk.IsPKrValid(&self.tx.From) {
		e = verify_utils.ReportError("txs.verify from is invalid", self.tx)
		return
	}
	if !c_superzk.VerifyPKr_X(&self.hash, &self.tx.Sign, &self.tx.From) {
		e = verify_utils.ReportError("txs.verify from verify failed", self.tx)
		return
	}
	return
}

func (self *verifyWithoutStateCtx) verifyPkg() (e error) {
	if self.tx.Desc_Cmd.Count() > 0 && self.tx.Desc_Pkg.Count() > 0 {
		e = verify_utils.ReportError("pkg and cmd desc only exists one", self.tx)
		return
	}
	if !self.tx.Desc_Pkg.Valid() {
		e = verify_utils.ReportError("pkg desc is invalid", self.tx)
		return
	}
	if self.tx.Desc_Pkg.Create != nil {
		self.zout_count++
	}
	if self.tx.Desc_Pkg.Close != nil {
		self.zin_count++
	}
	return
}

func (self *verifyWithoutStateCtx) verifyCmd() (e error) {
	if !self.tx.Desc_Cmd.Valid() {
		e = verify_utils.ReportError("cmd desc is invalid", self.tx)
		return
	}
	if asset := self.tx.Desc_Cmd.OutAsset(); asset != nil {
		self.oout_count++
		if asset.Tkn != nil {
			if !verify_utils.CheckUint(&asset.Tkn.Value) {
				e = verify_utils.ReportError("cmd asset tkn value invalid", self.tx)
				return
			}
		}
		self.tx.Desc_Cmd.ToAssetCC_Szk()
	}
	if pkr := self.tx.Desc_Cmd.ToPkr(); pkr != nil {
		if !superzk.IsPKrValid(pkr) {
			e = verify_utils.ReportError("cmd pkr invalid", self.tx)
			return
		}
	}
	if self.tx.Desc_Cmd.RegistPool != nil {
		if self.tx.Desc_Cmd.RegistPool.FeeRate > deceparam.HIGHEST_STAKING_NODE_FEE_RATE {
			e = verify_utils.ReportError(fmt.Sprintf("regist pool the fee rate must < %v%%", deceparam.HIGHEST_STAKING_NODE_FEE_RATE), self.tx)
			return
		}
		if self.tx.Desc_Cmd.RegistPool.FeeRate < deceparam.LOWEST_STAKING_NODE_FEE_RATE {
			e = verify_utils.ReportError(fmt.Sprintf("regist pool fee must >= %v%%", deceparam.LOWEST_STAKING_NODE_FEE_RATE/100), self.tx)
			return
		}
	}
	if self.tx.Desc_Cmd.Contract != nil {
		if self.tx.Desc_Cmd.Contract.To != nil {
			empty := c_type.PKr{}
			if *self.tx.Desc_Cmd.Contract.To == empty {
				e = verify_utils.ReportError("contract target can not be zero", self.tx)
				return
			}
		}
	}
	return
}

func (self *verifyWithoutStateCtx) verifyInsP() (e error) {
	self.oin_count += len(self.tx.Tx1.Ins_P)
	return
}

func (self *verifyWithoutStateCtx) verifyInsC() (e error) {
	for _, in := range self.tx.Tx1.Ins_C {
		if !c_superzk.VerifyZPKa(&self.hash, &in.Sign, &in.ZPKa) {
			e = verify_utils.ReportError("c_out zpka verify invalid", self.tx)
			return
		}
		self.zin_count++
	}
	return
}

func (self *verifyWithoutStateCtx) verifyOutP() (e error) {
	for i, out := range self.tx.Tx1.Outs_P {
		self.oout_count++
		if !superzk.IsPKrValid(&out.PKr) {
			e = verify_utils.ReportError("p_out pkr invalid", self.tx)
			return
		}
		if out.Asset.Tkn != nil {
			if !verify_utils.CheckUint(&out.Asset.Tkn.Value) {
				e = verify_utils.ReportError("p_out tkn value invalid", self.tx)
				return
			}
		}
		self.tx.Tx1.Outs_P[i].ToAssetCC_Szk()
	}
	return
}

func (self *verifyWithoutStateCtx) verifyOutC() (e error) {
	for _, out := range self.tx.Tx1.Outs_C {
		self.zout_count++
		if !c_superzk.IsPKrValid(&out.PKr) {
			e = verify_utils.ReportError("c_out pkr invalid", self.tx)
			return
		}
	}
	return
}

func (self *verifyWithoutStateCtx) verifyBalance() (e error) {
	if self.oout_count+self.zout_count > deceparam.MAX_Z_OUT_LENGTH_SIP2 {
		e = verify_utils.ReportError("verify error: out_size > 500", self.tx)
		return
	}
	return
}

func (self *verifyWithoutStateCtx) verify() (e error) {
	self.prepare()
	defer self.clear()

	self.ProcessVerifyProof()

	if e = self.verifyFee(); e != nil {
		return
	}
	if e = self.verifyFrom(); e != nil {
		return
	}
	if e = self.verifyPkg(); e != nil {
		return
	}
	if e = self.verifyCmd(); e != nil {
		return
	}
	if self.tx.Tx1.Count() > 0 {
		if e = self.verifyInsP(); e != nil {
			return
		}
		if e = self.verifyInsC(); e != nil {
			return
		}
		if e = self.verifyOutP(); e != nil {
			return
		}
		if e = self.verifyOutC(); e != nil {
			return
		}
	}
	if e = self.verifyBalance(); e != nil {
		return
	}
	if e = self.WaitVerifyProof(); e != nil {
		return
	}
	return
}
