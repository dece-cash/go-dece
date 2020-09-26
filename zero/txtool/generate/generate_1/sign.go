package generate_1

import (
	"errors"
	"github.com/dece-cash/go-dece/zero/txs/stx"

	"github.com/dece-cash/go-dece/zero/txs/pkg"

	"github.com/dece-cash/go-dece/core/types"

	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/superzk"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txs/assets"
	"github.com/dece-cash/go-dece/zero/txs/stx/tx"
	"github.com/dece-cash/go-dece/zero/txtool"
)

type sign_ctx struct {
	param        txtool.GTxParam
	p0_ins       []*txtool.GIn
	p_ins        []*txtool.GIn
	c_ins        []*txtool.GIn
	c_outs       []*txtool.GOut
	p_outs       []*txtool.GOut
	balance_desc c_type.BalanceDesc
	keys         []c_type.Uint256
	bases        []c_type.Uint256
	s            stx.T
	ck           assets.CKState
}

func (self *sign_ctx) Tx() (ret stx.T) {
	ret = self.s
	return
}

func (self *sign_ctx) Param() (ret txtool.GTxParam) {
	ret = self.param
	return
}

func (self *sign_ctx) Keys() (ret []c_type.Uint256) {
	ret = self.keys
	return
}

func (self *sign_ctx) Bases() (ret []c_type.Uint256) {
	ret = self.bases
	return
}

func SignTx(param *txtool.GTxParam) (ctx sign_ctx, e error) {
	ctx.param = *param
	if e = ctx.check(); e != nil {
		return
	}
	if e = ctx.prepare(); e != nil {
		return
	}
	if e = ctx.genFrom(); e != nil {
		return
	}
	if e = ctx.genFee(); e != nil {
		return
	}
	if e = ctx.genCmd(); e != nil {
		return
	}
	if e = ctx.genInsP(); e != nil {
		return
	}
	if e = ctx.genInsC(); e != nil {
		return
	}
	if e = ctx.genOutsC(); e != nil {
		return
	}
	if e = ctx.genOutsP(); e != nil {
		return
	}
	if e = ctx.genSign(); e != nil {
		return
	}
	return
}

func (self *sign_ctx) check() (e error) {
	sk := self.param.From.SKr.ToUint512()
	tk, e := superzk.Sk2Tk(&sk)
	if !superzk.IsMyPKr(&tk, &self.param.From.PKr) {
		e = errors.New("sk unmatch pkr for the From field")
		return
	}

	for _, in := range self.param.Ins {
		tk, _ := superzk.Sk2Tk(in.SKr.ToUint512().NewRef())

		if in.Out.State.OS.Out_P != nil {
			if !superzk.IsMyPKr(&tk, &in.Out.State.OS.Out_P.PKr) {
				e = errors.New("sk unmatch pkr for out_z")
				return
			}
			continue
		}
		if in.Out.State.OS.Out_C != nil {
			if !superzk.IsMyPKr(&tk, &in.Out.State.OS.Out_C.PKr) {
				e = errors.New("sk unmatch pkr for out_z")
				return
			}
			continue
		}
	}
	return
}

func (self *sign_ctx) prepare() (e error) {
	for i := range self.param.Ins {
		in := &self.param.Ins[i]
		if in.Out.State.OS.Out_P != nil {
			self.p_ins = append(self.p_ins, in)
			continue
		}
		if in.Out.State.OS.Out_C != nil {
			if self.param.Z != nil && *self.param.Z {
				self.c_ins = append(self.c_ins, in)
			} else {
				self.p_ins = append(self.p_ins, in)
			}
			continue
		}
	}
	for i := range self.param.Outs {
		out := &self.param.Outs[i]
		if c_superzk.IsSzkPKr(&out.PKr) {
			if self.param.Z != nil && *self.param.Z {
				self.c_outs = append(self.c_outs, out)
			} else {
				self.p_outs = append(self.p_outs, out)
			}
		} else {
			self.p_outs = append(self.p_outs, out)
		}
	}
	self.s.Ehash = types.Ehash(*self.param.GasPrice, self.param.Gas, []byte{})
	return
}

func (self *sign_ctx) genFrom() (e error) {
	self.s.From = self.param.From.PKr
	return
}

func (self *sign_ctx) genFee() (e error) {
	self.s.Fee = self.param.Fee

	if cc, err := c_superzk.GenAssetCC(self.param.Fee.ToTypeAsset().NewRef()); err != nil {
		e = err
		return
	} else {
		self.ck = assets.NewCKState(true, &self.param.Fee)
		self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, cc[:]...)
	}
	return
}

func (self *sign_ctx) genCmd() (e error) {
	var a *assets.Asset
	if self.param.Cmds.BuyShare != nil {
		self.s.Desc_Cmd.BuyShare = self.param.Cmds.BuyShare
		asset := self.param.Cmds.BuyShare.Asset()
		a = &asset
	}
	if self.param.Cmds.RegistPool != nil {
		self.s.Desc_Cmd.RegistPool = self.param.Cmds.RegistPool
		asset := self.param.Cmds.RegistPool.Asset()
		a = &asset
	}
	if self.param.Cmds.ClosePool != nil {
		self.s.Desc_Cmd.ClosePool = self.param.Cmds.ClosePool
	}
	if self.param.Cmds.Contract != nil {
		self.s.Desc_Cmd.Contract = self.param.Cmds.Contract
		a = &self.param.Cmds.Contract.Asset
	}
	if self.param.Cmds.PkgCreate != nil {
		create := self.param.Cmds.PkgCreate
		self.s.Desc_Pkg.Create = &stx.PkgCreate{}
		self.s.Desc_Pkg.Create.PKr = create.PKr
		self.s.Desc_Pkg.Create.Id = create.Id
		create.Ar = c_superzk.RandomFr()
		if cm, _, err := c_superzk.GenAssetCM_PC(create.Asset.ToTypeAsset().NewRef(), &create.Ar); err != nil {
			e = err
			return
		} else {
			self.s.Desc_Pkg.Create.Pkg.AssetCM = cm
		}
		sk := self.param.From.SKr.ToUint512()
		tk, err := superzk.Sk2Tk(&sk)
		if err != nil {
			e = err
			return
		}
		key := pkg.GetKey(&self.param.From.PKr, &tk)
		if einfo, err := c_superzk.EncInfo(&key, create.Asset.ToTypeAsset().NewRef(), &create.Memo, &create.Ar); err != nil {
			e = err
			return
		} else {
			self.s.Desc_Pkg.Create.Pkg.EInfo = einfo
		}
	}
	if self.param.Cmds.PkgTransfer != nil {
		change := self.param.Cmds.PkgTransfer
		self.s.Desc_Pkg.Transfer = &stx.PkgTransfer{}
		self.s.Desc_Pkg.Transfer.Id = change.Id
		self.s.Desc_Pkg.Transfer.PKr = change.PKr
	}
	if self.param.Cmds.PkgClose != nil {
		close := self.param.Cmds.PkgClose
		self.s.Desc_Pkg.Close = &stx.PkgClose{}
		self.s.Desc_Pkg.Close.Id = close.Id
		self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, close.AssetCM[:]...)
		self.balance_desc.Zin_ars = append(self.balance_desc.Zin_ars, close.Ar[:]...)
	}
	if a != nil {
		if cc, err := c_superzk.GenAssetCC(a.ToTypeAsset().NewRef()); err != nil {
			e = err
			return
		} else {
			self.ck.AddOut(a)
			self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, cc[:]...)
		}
	}
	return
}

func (self *sign_ctx) genInsP() (e error) {
	for _, in := range self.p_ins {
		sk := in.SKr.ToUint512()
		tk, _ := superzk.Sk2Tk(&sk)

		t_in := tx.In_P{}
		t_in.Root = in.Out.Root
		t_in.Nil, e = c_superzk.GenNil(&tk, in.Out.State.OS.RootCM, in.Out.State.OS.ToPKr())
		if e != nil {
			return
		}
		var asset_desc assets.Asset
		if in.Out.State.OS.Out_P != nil {
			asset_desc = in.Out.State.OS.Out_P.Asset
		} else {
			if out_c := in.Out.State.OS.Out_C; out_c != nil {
				if key, _, err := c_superzk.FetchKey(&out_c.PKr, &tk, &out_c.RPK); err != nil {
					e = err
					return
				} else {
					if dout, _ := ConfirmOutC(&key, out_c); dout == nil {
						e = errors.New("gen tx1 confirm outz error")
						return
					} else {
						t_in.Key = &key
						asset_desc = dout.Asset
					}
				}
			} else {
				e = errors.New("gen in_p but no out_p or out_c")
			}
		}

		if cc, err := c_superzk.GenAssetCC(asset_desc.ToTypeAsset().NewRef()); err != nil {
			e = err
			return
		} else {
			self.ck.AddIn(&asset_desc)
			self.balance_desc.Oin_accs = append(self.balance_desc.Oin_accs, cc[:]...)
			self.s.Tx1.Ins_P = append(self.s.Tx1.Ins_P, t_in)
		}
	}
	return
}

func (self *sign_ctx) genInsC() (e error) {
	for _, in := range self.c_ins {
		sk := in.SKr.ToUint512()
		tk, _ := superzk.Sk2Tk(&sk)

		t_in := tx.In_C{}

		t_in.Nil, e = c_superzk.GenNil(&tk, in.Out.State.OS.RootCM, in.Out.State.OS.ToPKr())
		if e != nil {
			return
		}

		key, vskr, err := c_superzk.FetchKey(&in.Out.State.OS.Out_C.PKr, &tk, &in.Out.State.OS.Out_C.RPK)
		if err != nil {
			e = err
			return
		}
		in.Vskr = &vskr

		dout, ar_old := ConfirmOutC(&key, in.Out.State.OS.Out_C)
		if dout == nil {
			e = errors.New("gen in_c error: can not find out_c")
			return
		}
		in.ArOld = &ar_old
		// self.keys = append(self.keys, key)

		in.Ar = c_superzk.RandomFr().NewRef()
		cm, cc, err := c_superzk.GenAssetCM_PC(dout.Asset.ToTypeAsset().NewRef(), in.Ar)
		if err != nil {
			e = err
			return
		}
		t_in.AssetCM = cm
		in.CC = &cc

		in.A = c_superzk.RandomFr().NewRef()
		t_in.ZPKa, e = c_superzk.GenZPKa(&in.Out.State.OS.Out_C.PKr, in.A)
		if e != nil {
			return
		}

		t_in.Anchor = in.Witness.Anchor

		self.balance_desc.Zin_acms = append(self.balance_desc.Zin_acms, t_in.AssetCM[:]...)
		self.balance_desc.Zin_ars = append(self.balance_desc.Zin_ars, in.Ar[:]...)
		self.s.Tx1.Ins_C = append(self.s.Tx1.Ins_C, t_in)

		baser := c_superzk.ClearPKr(in.Out.State.OS.ToPKr()).BASEr()
		self.bases = append(self.bases, baser)
	}
	return
}

func (self *sign_ctx) genOutsC() (e error) {
	for _, out := range self.c_outs {
		t_out := tx.Out_C{}

		out.Ar = c_superzk.RandomFr().NewRef()
		t_out.AssetCM, _, e = c_superzk.GenAssetCM_PC(out.Asset.ToTypeAsset().NewRef(), out.Ar)
		if e != nil {
			return
		}

		t_out.PKr = out.PKr
		var key c_type.Uint256
		key, t_out.RPK, _, e = c_superzk.GenKey(&out.PKr)
		if e != nil {
			return
		}

		t_out.EInfo, e = c_superzk.EncInfo(&key, out.Asset.ToTypeAsset().NewRef(), &out.Memo, out.Ar)
		if e != nil {
			return
		}

		self.balance_desc.Zout_acms = append(self.balance_desc.Zout_acms, t_out.AssetCM[:]...)
		self.balance_desc.Zout_ars = append(self.balance_desc.Zout_ars, out.Ar[:]...)
		self.s.Tx1.Outs_C = append(self.s.Tx1.Outs_C, t_out)
		self.keys = append(self.keys, key)
	}
	return
}

func (self *sign_ctx) genOutsP() (e error) {
	for _, out := range self.p_outs {
		t_out := tx.Out_P{}
		t_out.PKr = out.PKr
		t_out.Asset = out.Asset
		t_out.Memo = out.Memo

		if cc, err := c_superzk.GenAssetCC(out.Asset.ToTypeAsset().NewRef()); err != nil {
			e = err
			return
		} else {
			self.ck.AddOut(&out.Asset)
			self.balance_desc.Oout_accs = append(self.balance_desc.Oout_accs, cc[:]...)
			self.s.Tx1.Outs_P = append(self.s.Tx1.Outs_P, t_out)
		}
	}
	return
}

func (self *sign_ctx) genSign() (e error) {

	self.balance_desc.Hash = self.s.Tx1_Hash()

	if e = self.signFrom(); e != nil {
		return
	}
	if e = self.signInsP(); e != nil {
		return
	}
	if e = self.signInsC(); e != nil {
		return
	}
	if e = self.signPkg(); e != nil {
		return
	}
	if e = self.signBalance(); e != nil {
		return
	}
	return
}

func (self *sign_ctx) signFrom() (e error) {
	if sign, err := c_superzk.SignPKr_X(self.param.From.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, &self.s.From); err != nil {
		return err
	} else {
		self.s.Sign = sign
		return nil
	}
	return
}

func (self *sign_ctx) signInsP() (e error) {
	for i := range self.s.Tx1.Ins_P {
		t_in := self.p_ins[i]
		if sign, err := c_superzk.SignPKr_P(t_in.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, t_in.Out.State.OS.ToPKr()); err != nil {
			return err
		} else {
			self.s.Tx1.Ins_P[i].ASign = sign
		}
		tk, _ := superzk.Sk2Tk(t_in.SKr.ToUint512().NewRef())
		if sign, err := c_superzk.SignNil(
			&tk,
			&self.balance_desc.Hash,
			t_in.Out.State.OS.RootCM.NewRef(),
			t_in.Out.State.OS.ToPKr(),
		); err != nil {
			e = err
			return
		} else {
			self.s.Tx1.Ins_P[i].NSign = sign
		}
	}
	return
}
func (self *sign_ctx) signInsC() (e error) {
	for i := range self.s.Tx1.Ins_C {
		t_in := self.c_ins[i]
		if sign, err := c_superzk.SignZPKa(t_in.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, t_in.A, t_in.Out.State.OS.ToPKr()); err != nil {
			e = err
			return
		} else {
			self.s.Tx1.Ins_C[i].Sign = sign
		}
	}
	return
}

func (self *sign_ctx) signPkg() error {
	if self.param.Cmds.PkgTransfer != nil {
		if sign, err := c_superzk.SignPKr_X(self.param.From.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, &self.param.Cmds.PkgTransfer.Owner); err != nil {
			return err
		} else {
			self.s.Desc_Pkg.Transfer.Sign = sign
		}
	}
	if self.param.Cmds.PkgClose != nil {
		if sign, err := c_superzk.SignPKr_X(self.param.From.SKr.ToUint512().NewRef(), &self.balance_desc.Hash, &self.param.Cmds.PkgClose.Owner); err != nil {
			return err
		} else {
			self.s.Desc_Pkg.Transfer.Sign = sign
		}
	}
	return nil
}

func (self *sign_ctx) signBalance() (e error) {
	if len(self.balance_desc.Zin_acms) > 0 || len(self.balance_desc.Zout_acms) > 0 {
		if e = c_superzk.SignBalance(&self.balance_desc); e != nil {
			return
		}
		if self.balance_desc.Bcr == c_type.Empty_Uint256 {
			return errors.New("sign balance failed!!!")
		} else {
			self.s.Bcr = self.balance_desc.Bcr
			self.s.Bsign = self.balance_desc.Bsign
			return nil
		}
	} else {
		if e = self.ck.Check(); e != nil {
			return
		} else {
			return
		}
	}
}
