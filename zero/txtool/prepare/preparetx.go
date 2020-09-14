package prepare

import (
	"bytes"

	"github.com/dece-cash/go-dece/czero/superzk"

	"github.com/dece-cash/go-dece/zero/utils"

	"github.com/dece-cash/go-dece/zero/txtool"

	"github.com/pkg/errors"
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/common"
)

func GenTxParam(param *PreTxParam, gen TxParamGenerator, state TxParamState) (txParam *txtool.GTxParam, e error) {
	if len(param.Receptions) > 500 {
		return nil, errors.New("receptions count must <= 500")
	}
	utxos, err := SelectUtxos(param, gen)
	if err != nil {
		return nil, err
	}

	if param.RefundTo == nil {
		if param.RefundTo = gen.DefaultRefundTo(&param.From); param.RefundTo == nil {
			return nil, errors.New("can not find default refund to")
		}
	}
	bparam := BeforeTxParam{
		param.Fee,
		*param.GasPrice,
		utxos,
		*param.RefundTo,
		param.Receptions,
		param.Cmds,
	}
	txParam, e = BuildTxParam(state, &bparam)
	return
}

func IsPk(addr c_type.PKr) bool {
	byte32 := common.Hash{}
	return bytes.Equal(byte32[:], addr[64:96])
}

func CreatePkr(pk *c_type.Uint512, index uint64) c_type.PKr {
	r := c_type.Uint256{}
	copy(r[:], common.LeftPadBytes(utils.EncodeNumber(index), 32))
	if index == 0 {
		return superzk.Pk2PKr(pk, nil)
	} else {
		return superzk.Pk2PKr(pk, &r)
	}
}
