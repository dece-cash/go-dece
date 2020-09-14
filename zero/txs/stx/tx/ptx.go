package tx

import (
	"sync/atomic"

	"github.com/dece-cash/go-dece/czero/c_superzk"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto/sha3"
	"github.com/dece-cash/go-dece/zero/txs/assets"
	"github.com/dece-cash/go-dece/zero/utils"
)

type In_P struct {
	Root  c_type.Uint256
	Nil   c_type.Uint256
	Key   *c_type.Uint256 `rlp:"nil"`
	NSign c_type.SignN
	ASign c_type.Uint512
}

func (self *In_P) Tx1_Hash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Root[:])
	d.Write(self.Nil[:])
	if self.Key != nil {
		d.Write(self.Key[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *In_P) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.Root[:])
	d.Write(self.Nil[:])
	if self.Key != nil {
		d.Write(self.Key[:])
	}
	d.Write(self.NSign[:])
	d.Write(self.ASign[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

type Out_P struct {
	PKr   c_type.PKr
	Asset assets.Asset
	Memo  c_type.Uint512

	assetCC_Szk atomic.Value
}

func (self *Out_P) ToAssetCC_Szk() c_type.Uint256 {
	if cc, ok := self.assetCC_Szk.Load().(c_type.Uint256); ok {
		return cc
	}
	v, _ := c_superzk.GenAssetCC(self.Asset.ToTypeAsset().NewRef())
	self.assetCC_Szk.Store(v)
	return v
}

func (self *Out_P) Clone() (ret Out_P) {
	utils.DeepCopy(&ret, self)
	return
}

func (this Out_P) ToRef() (ret *Out_P) {
	ret = &this
	return
}

func (self *Out_P) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.PKr[:])
	d.Write(self.Asset.ToHash().NewRef()[:])
	d.Write(self.Memo[:])
	copy(ret[:], d.Sum(nil))
	return ret
}
