package localdb

import (
	"github.com/pkg/errors"
	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto"
	"github.com/dece-cash/go-dece/zero/utils"
)

func HashIndexRsk(index uint64) (ret c_type.Uint256) {
	ret = utils.NewU256(index).ToRef().ToUint256()
	return
}

func HashIndexAr(index uint64) (ret c_type.Uint256) {
	index_bytes := utils.NewU256(index)
	pre_ar := index_bytes.ToUint256()
	ar_bytes := crypto.Keccak256(pre_ar[:])
	copy(ret[:], ar_bytes)
	ret = c_superzk.ForceFr(&ret)
	return
}


func genRootCM(self *OutState) (cm c_type.Uint256, e error) {
	if self.Out_P != nil {
		ar := HashIndexAr(self.Index)
		type_asset := self.Out_P.Asset.ToTypeAsset()
		cm, e = c_superzk.GenRootCM_P(
			self.Index,
			&type_asset,
			&ar,
			&self.Out_P.PKr,
		)
		return
	} else if self.Out_C != nil {
		cm, e = c_superzk.GenRootCM_C(
			self.Index,
			&self.Out_C.AssetCM,
			&self.Out_C.PKr,
		)
		return
	} else {
		e = errors.New("no output for root cm")
		return
	}
}
