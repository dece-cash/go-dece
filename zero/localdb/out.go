package localdb

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/rlp"
	"github.com/dece-cash/go-dece/zero/txs/stx/tx"
	"github.com/dece-cash/go-dece/zero/utils"
)

type OutState struct {
	Index  uint64
	Out_P  *tx.Out_P       `rlp:"nil"`
	Out_C  *tx.Out_C       `rlp:"nil"`
	OutCM  *c_type.Uint256 `rlp:"nil"`
	RootCM *c_type.Uint256 `rlp:"nil"`
}

func (self *OutState) GenRootCM() {
	if self.RootCM == nil {
		if cm, err := genRootCM(self); err != nil {
			panic(err)
			return
		} else {
			self.RootCM = &cm
			return
		}
	} else {
		return
	}
}

// func (b *OutState) DecodeRLP(s *rlp.Stream) error {
// 	if e := s.Decode(b); e != nil {
// 		return e
// 	}
// 	return nil
// }
//
// func (b *OutState) EncodeRLP(w io.Writer) error {
// 	return rlp.Encode(w, b)
// }

func (out *OutState) TxType() string {
	if out.Out_P != nil {
		return "Out_P"
	}
	if out.Out_C != nil {
		return "Out_C"
	}
	return "EMPTY"
}

func (out *OutState) IsZero() bool {
	if out.Out_C != nil {
		return true
	} else {
		return false
	}
}
func (out *OutState) IsSzk() bool {
	if out.Out_P != nil || out.Out_C != nil {
		return true
	}
	return false
}

func (self *OutState) Clone() (ret OutState) {
	utils.DeepCopy(&ret, self)
	return
}

func (self *OutState) ToPKr() *c_type.PKr {
	if self.Out_P != nil {
		return &self.Out_P.PKr
	} else if self.Out_C != nil {
		return &self.Out_C.PKr
	}
	return nil
}

func (self *OutState) Serial() (ret []byte, e error) {
	if self != nil {
		return rlp.EncodeToBytes(self)
	} else {
		return
	}
}

type OutState0Get struct {
	Out *OutState
}

func (self *OutState0Get) Unserial(v []byte) (e error) {
	if len(v) == 0 {
		self.Out = nil
		return
	} else {
		self.Out = &OutState{}
		if err := rlp.DecodeBytes(v, &self.Out); err != nil {
			e = err
			return
		} else {
			return
		}
	}
}
