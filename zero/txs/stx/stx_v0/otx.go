// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package stx_v0

import (

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto/sha3"
	"github.com/dece-cash/go-dece/zero/utils"

	"github.com/dece-cash/go-dece/zero/txs/assets"

	"github.com/dece-cash/go-dece/crypto"
)

type Out_O struct {
	Addr  c_type.PKr
	Asset assets.Asset
	Memo  c_type.Uint512
}

func (self *Out_O) Clone() (ret Out_O) {
	utils.DeepCopy(&ret, self)
	return
}
func (this Out_O) ToRef() (ret *Out_O) {
	ret = &this
	return
}

func (self *Out_O) ToHash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Addr[:],
		self.Asset.ToHash().NewRef()[:],
		self.Memo[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *Out_O) ToHash_for_gen() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Addr[:],
		self.Asset.ToHash().NewRef()[:],
		self.Memo[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *Out_O) ToHash_for_sign() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Addr[:],
		self.Asset.ToHash().NewRef()[:],
		self.Memo[:],
	)
	copy(ret[:], hash)
	return ret
}

type In_S struct {
	Root c_type.Uint256
	Nil  c_type.Uint256
	Sign c_type.Uint512
}

func (self *In_S) ToHash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Root[:],
		self.Nil[:],
		self.Sign[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *In_S) ToHash_for_gen() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Root[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *In_S) ToHash_for_sign() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Root[:],
	)
	copy(ret[:], hash)
	return ret
}

type Desc_O struct {
	Ins  []In_S
	Outs []Out_O
}

func (self *Desc_O) HasContent() bool {
	return len(self.Ins) > 0 || len(self.Outs) > 0
}

func (self *Desc_O) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Desc_O) ToHash_for_gen() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash_for_gen().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash_for_gen().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}

func (self *Desc_O) ToHash_for_sign() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	for _, in := range self.Ins {
		d.Write(in.ToHash_for_sign().NewRef()[:])
	}
	for _, out := range self.Outs {
		d.Write(out.ToHash_for_sign().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return
}
