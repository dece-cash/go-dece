package tx

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto"
	"github.com/dece-cash/go-dece/zero/utils"
)

type Out_C struct {
	PKr     c_type.PKr
	AssetCM c_type.Uint256
	RPK     c_type.Uint256
	EInfo   c_type.Einfo
	Proof   c_type.Proof
}

func (self *Out_C) Clone() (ret Out_C) {
	utils.DeepCopy(&ret, self)
	return
}

func (self *Out_C) Tx1_Hash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.PKr[:],
		self.AssetCM[:],
		self.RPK[:],
		self.EInfo[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *Out_C) ToHash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.PKr[:],
		self.AssetCM[:],
		self.RPK[:],
		self.EInfo[:],
		self.Proof[:],
	)
	copy(ret[:], hash)
	return ret
}

type In_C struct {
	Anchor  c_type.Uint256
	Nil     c_type.Uint256
	AssetCM c_type.Uint256
	ZPKa    c_type.Uint256
	Sign    c_type.Uint512
	Proof   c_type.Proof
}

func (self *In_C) Tx1_Hash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Anchor[:],
		self.Nil[:],
		self.AssetCM[:],
		self.ZPKa[:],
	)
	copy(ret[:], hash)
	return ret
}

func (self *In_C) ToHash() (ret c_type.Uint256) {
	hash := crypto.Keccak256(
		self.Anchor[:],
		self.Nil[:],
		self.AssetCM[:],
		self.ZPKa[:],
		self.Sign[:],
		self.Proof[:],
	)
	copy(ret[:], hash)
	return ret
}
