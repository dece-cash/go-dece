package assets

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto/sha3"
	"github.com/dece-cash/go-dece/zero/utils"
)

type Asset struct {
	Tkn *Token  `rlp:"nil"`
	Tkt *Ticket `rlp:"nil"`
}

func (self *Asset) IsValid() bool {
	if self.Tkn != nil {
		return self.Tkn.IsValid()
	}
	return true
}

func (self *Asset) HasAsset() bool {
	if self != nil {
		if self.Tkn != nil {
			if self.Tkn.Value.Cmp(&utils.U256_0) != 0 {
				return true
			}
		}
		if self.Tkt != nil {
			if self.Tkt.Value != c_type.Empty_Uint256 {
				return true
			}
		}
	}
	return false
}

func NewAssetByType(asset *c_type.Asset) (ret Asset) {
	ret = NewAsset(
		&Token{
			asset.Tkn_currency,
			utils.NewU256_ByKey(&asset.Tkn_value),
		},
		&Ticket{
			asset.Tkt_category,
			asset.Tkt_value,
		},
	)
	return
}
func NewAsset(tkn *Token, tkt *Ticket) (ret Asset) {
	if tkn != nil {
		if tkn.Value.Cmp(&utils.U256_0) > 0 && tkn.Currency != c_type.Empty_Uint256 {
			ret.Tkn = tkn.Clone().ToRef()
		}
	}
	if tkt != nil {
		if tkt.Value != c_type.Empty_Uint256 && tkt.Category != c_type.Empty_Uint256 {
			ret.Tkt = tkt.Clone().ToRef()
		}
	}
	return
}

func (self Asset) ToRef() (ret *Asset) {
	return &self
}

func (self *Asset) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	if self.Tkn != nil {
		d.Write(self.Tkn.ToHash().NewRef()[:])
	}
	if self.Tkt != nil {
		d.Write(self.Tkt.ToHash().NewRef()[:])
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *Asset) Clone() (ret Asset) {
	utils.DeepCopy(&ret, self)
	return
}
