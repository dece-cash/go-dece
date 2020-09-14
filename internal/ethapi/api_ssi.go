package ethapi

import (
	"context"

	"github.com/dece-cash/go-dece/zero/wallet/ssi"

	"github.com/dece-cash/go-dece/zero/txtool"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/common/hexutil"
)

type PublicSSIAPI struct {
	b Backend
}

func (s *PublicSSIAPI) SzkCreateKr() (kr txtool.Kr) {
	return ssi.SSI_Inst.CreateKr(true)
}

func (s *PublicSSIAPI) CreateKr() (kr txtool.Kr) {
	return ssi.SSI_Inst.CreateKr(false)
}

func (s *PublicSSIAPI) GetBlocksInfo(ctx context.Context, start hexutil.Uint64, count hexutil.Uint64) ([]ssi.Block, error) {
	return ssi.SSI_Inst.GetBlocksInfo(uint64(start), uint64(count))
}

func (s *PublicSSIAPI) Detail(ctx context.Context, roots []c_type.Uint256, skr *c_type.PKr) (douts []txtool.DOut, e error) {
	return ssi.SSI_Inst.Detail(roots, skr)
}

func (s *PublicSSIAPI) GenTx(ctx context.Context, param *ssi.PreTxParam) (hash c_type.Uint256, e error) {
	return ssi.SSI_Inst.GenTx(param)
}

func (s *PublicSSIAPI) GetTx(ctx context.Context, txhash c_type.Uint256) (tx *txtool.GTx, e error) {
	return ssi.SSI_Inst.GetTx(txhash)
}

func (s *PublicSSIAPI) CommitTx(ctx context.Context, txhash c_type.Uint256) (e error) {
	if tx, err := ssi.SSI_Inst.GetTx(txhash); err != nil {
		e = err
		return
	} else {
		return s.b.CommitTx(tx)
	}
}
