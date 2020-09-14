package localdb

import (
	"math/big"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto/sha3"
	"github.com/dece-cash/go-dece/rlp"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/txs/stx"
	"github.com/dece-cash/go-dece/zero/txs/zstate/tri"
)

type ZPkg struct {
	High   uint64
	From   c_type.PKr
	Pack   stx.PkgCreate
	Closed bool
}

func (self *ZPkg) ToHash() (ret c_type.Uint256) {
	d := sha3.NewKeccak256()
	d.Write(big.NewInt(int64(self.High)).Bytes())
	d.Write(self.From[:])
	d.Write(self.Pack.ToHash().NewRef()[:])
	if self.Closed {
		d.Write([]byte{1})
	} else {
		d.Write([]byte{0})
	}
	copy(ret[:], d.Sum(nil))
	return ret
}

func (self *ZPkg) Serial() (ret []byte, e error) {
	return rlp.EncodeToBytes(self)
}

type PkgGet struct {
	Out *ZPkg
}

func (self *PkgGet) Unserial(v []byte) (e error) {
	if len(v) < 2 {
		self.Out = nil
		return
	} else {
		self.Out = &ZPkg{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			self.Out = nil
			return
		} else {
			return
		}
	}
}

func PkgKey(root *c_type.Uint256) []byte {
	key := []byte("$DECE_LOCALDB_PKG_HASH$")
	key = append(key, root[:]...)
	return key
}

func PutPkg(db decedb.Putter, hash *c_type.Uint256, pkg *ZPkg) {
	key := PkgKey(hash)
	tri.UpdateDBObj(db, key, pkg)
}

func GetPkg(db decedb.Getter, hash *c_type.Uint256) (ret *ZPkg) {
	key := PkgKey(hash)
	get := PkgGet{}
	tri.GetDBObj(db, key, &get)
	ret = get.Out
	return
}
