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

package stx

import (
	"sync/atomic"

	"github.com/dece-cash/go-dece/czero/c_superzk"

	"github.com/dece-cash/go-dece/crypto/sha3"
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txs/assets"
	"github.com/dece-cash/go-dece/zero/txs/stx/tx"
)

type Tx2 struct {
}

type T struct {
	Ehash    c_type.Uint256
	From     c_type.PKr
	Fee      assets.Token
	Sign     c_type.Uint512
	Bcr      c_type.Uint256
	Bsign    c_type.Uint512
	Desc_Pkg PkgDesc_Z
	Desc_Cmd DescCmd
	Tx       tx.Tx

	// cache
	hash      atomic.Value
	feeCC_Szk atomic.Value
}

func (self *T) ContractAsset() *assets.Asset {
	if self.Desc_Cmd.Contract != nil {
		return &self.Desc_Cmd.Contract.Asset
	}
	return nil
}

func (self *T) ContractAddress() *c_type.PKr {
	if self.Desc_Cmd.Contract != nil {
		return self.Desc_Cmd.Contract.To
	}
	return nil
}

func (self *T) IsOpContract() bool {
	if self.Desc_Cmd.Contract != nil {
		return true
	}
	return false
}

func (self *T) ToFeeCC_Szk() c_type.Uint256 {
	if cc, ok := self.feeCC_Szk.Load().(c_type.Uint256); ok {
		return cc
	}
	v, _ := c_superzk.GenAssetCC(self.Fee.ToTypeAsset().NewRef())
	self.feeCC_Szk.Store(v)
	return v
}

func (self *T) ToHash() (ret c_type.Uint256) {
	if h, ok := self.hash.Load().(c_type.Uint256); ok {
		ret = h
		return
	}
	v := self._ToHash()
	self.hash.Store(v)
	return v
}

func (self *T) _ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	if self.Tx.Count() > 0 {
		d.Write(self.Tx.ToHash().NewRef()[:])
	}
	d.Write(self.Desc_Pkg.ToHash().NewRef()[:])
	d.Write(self.Sign[:])
	d.Write(self.Bcr[:])
	d.Write(self.Bsign[:])
	if self.Desc_Cmd.Count() > 0 {
		d.Write(self.Desc_Cmd.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash_for_gen() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	d.Write(self.Desc_Pkg.ToHash_for_gen().NewRef()[:])
	if self.Desc_Cmd.Count() > 0 {
		d.Write(self.Desc_Cmd.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *T) ToHash_for_sign() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Ehash[:])
	d.Write(self.From[:])
	d.Write(self.Fee.ToHash().NewRef()[:])
	d.Write(self.Desc_Pkg.ToHash_for_sign().NewRef()[:])
	if self.Desc_Cmd.Count() > 0 {
		d.Write(self.Desc_Cmd.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}
