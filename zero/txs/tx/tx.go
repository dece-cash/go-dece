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

package tx

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txs/assets"
	"github.com/dece-cash/go-dece/zero/txs/pkg"
	"github.com/dece-cash/go-dece/zero/utils"
)

type In struct {
	Root c_type.Uint256
	IsO  bool
}

type Out struct {
	Addr  c_type.PKr
	Asset assets.Asset
	Memo  c_type.Uint512
	IsZ   bool
}

type PkgCreate struct {
	Id  c_type.Uint256
	PKr c_type.PKr
	Pkg pkg.Pkg_O
}

type PkgClose struct {
	Id  c_type.Uint256
	Key c_type.Uint256
}

type PkgTransfer struct {
	Id  c_type.Uint256
	PKr c_type.PKr
}

type T struct {
	FromRnd     *c_type.Uint256
	Ehash       c_type.Uint256
	Fee         assets.Token
	Ins         []In
	Outs        []Out
	PkgCreate   *PkgCreate
	PkgTransfer *PkgTransfer
	PkgClose    *PkgClose
}

func (self *T) TokenCost() (ret map[c_type.Uint256]utils.U256) {
	ret = make(map[c_type.Uint256]utils.U256)
	ret[self.Fee.Currency] = self.Fee.Value
	if len(self.Outs) > 0 {
		for _, out := range self.Outs {
			if out.Asset.Tkn != nil {
				if cost, ok := ret[out.Asset.Tkn.Currency]; ok {
					cost.AddU(&out.Asset.Tkn.Value)
					ret[out.Asset.Tkn.Currency] = cost
				} else {
					ret[out.Asset.Tkn.Currency] = out.Asset.Tkn.Value
				}
			}
		}
	}
	if self.PkgCreate != nil {
		asset := self.PkgCreate.Pkg.Asset
		if asset.Tkn != nil {
			if cost, ok := ret[asset.Tkn.Currency]; ok {
				cost.AddU(&asset.Tkn.Value)
				ret[asset.Tkn.Currency] = cost
			} else {
				ret[asset.Tkn.Currency] = asset.Tkn.Value
			}
		}
	}
	return
}

func (self *T) TikectCost() (ret map[c_type.Uint256][]c_type.Uint256) {
	ret = make(map[c_type.Uint256][]c_type.Uint256)
	if len(self.Outs) > 0 {
		for _, out := range self.Outs {
			if out.Asset.Tkt != nil {
				if tkts, ok := ret[out.Asset.Tkt.Category]; ok {
					tkts = append(tkts, out.Asset.Tkt.Value)
					ret[out.Asset.Tkt.Category] = tkts
				} else {
					ret[out.Asset.Tkt.Category] = []c_type.Uint256{out.Asset.Tkt.Value}
				}
			}
		}
	}
	if self.PkgCreate != nil {
		asset := self.PkgCreate.Pkg.Asset
		if asset.Tkt != nil {
			if tkts, ok := ret[asset.Tkt.Category]; ok {
				tkts = append(tkts, asset.Tkt.Value)
				ret[asset.Tkt.Category] = tkts
			} else {
				ret[asset.Tkt.Category] = []c_type.Uint256{asset.Tkt.Value}
			}
		}
	}
	return
}
