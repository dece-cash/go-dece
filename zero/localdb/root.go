package localdb

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/rlp"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/txs/zstate/tri"
)

type RootState struct {
	OS     OutState
	TxHash c_type.Uint256
	Num    uint64
}

func (self *RootState) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type RootStateGet struct {
	Out *RootState
}

func (self *RootStateGet) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.Out = nil
		return
	} else {
		self.Out = &RootState{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}

func Root2TxHashKey(root *c_type.Uint256) []byte {
	key := []byte("$DECE_LOCALDB_ROOTSTATE$")
	key = append(key, root[:]...)
	return key
}

func RootCM2RootKey(root_cm *c_type.Uint256) []byte {
	key := []byte("$DECE_LOCALDB_ROOTCM2ROOT$")
	key = append(key, root_cm[:]...)
	return key
}

func PutRoot(db decedb.Putter, root *c_type.Uint256, rs *RootState) {
	rootkey := Root2TxHashKey(root)
	tri.UpdateDBObj(db, rootkey, rs)
	rootcmkey := RootCM2RootKey(rs.OS.RootCM)
	db.Put(rootcmkey, root[:])
}

func GetRoot(db decedb.Getter, root *c_type.Uint256) (ret *RootState) {
	rootkey := Root2TxHashKey(root)
	rootget := RootStateGet{}
	tri.GetDBObj(db, rootkey, &rootget)
	ret = rootget.Out
	return
}

func GetRootByRootCM(db decedb.Getter, root_cm *c_type.Uint256) (root *c_type.Uint256) {
	rootcmkey := RootCM2RootKey(root_cm)
	if root_bs, err := db.Get(rootcmkey); err != nil {
		return
	} else {
		root = &c_type.Uint256{}
		copy(root[:], root_bs)
		return
	}
}
