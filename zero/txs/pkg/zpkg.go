package pkg

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto/sha3"
	"github.com/dece-cash/go-dece/zero/utils"
)

type Pkg_Z struct {
	AssetCM c_type.Uint256
	EInfo   c_type.Einfo
}

func (this Pkg_Z) ToRef() (ret *Pkg_Z) {
	ret = &this
	return
}

func (self *Pkg_Z) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(self.AssetCM[:])
	d.Write(self.EInfo[:])
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Pkg_Z) Clone() (ret Pkg_Z) {
	utils.DeepCopy(&ret, self)
	return
}
