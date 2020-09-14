package types

import (
	"io"

	"github.com/dece-cash/go-dece/core/types/vserial"

	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/rlp"
)

type receiptForStorage_Version_0 struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	Bloom             Bloom
	TxHash            common.Hash
	ContractAddress   common.Address
	Logs              []*LogForStorage
	GasUsed           uint64
}

type receiptForStorage_Version_1 struct {
	PoolId  *common.Hash `rlp:"nil"`
	ShareId *common.Hash `rlp:"nil"`
}

// ReceiptForStorage is a wrapper around a Receipt that flattens and parses the
// entire content of a receipt, as opposed to only the consensus fields originally.
type ReceiptForStorage Receipt

// EncodeRLP implements rlp.Encoder, and flattens all content fields of a receipt
// into an RLP stream.
func (r *ReceiptForStorage) EncodeRLP(w io.Writer) error {
	vs := vserial.NewVSerial()
	{
		v0 := &receiptForStorage_Version_0{
			PostStateOrStatus: (*Receipt)(r).statusEncoding(),
			CumulativeGasUsed: r.CumulativeGasUsed,
			Bloom:             r.Bloom,
			TxHash:            r.TxHash,
			ContractAddress:   r.ContractAddress,
			Logs:              make([]*LogForStorage, len(r.Logs)),
			GasUsed:           r.GasUsed,
		}
		for i, log := range r.Logs {
			v0.Logs[i] = (*LogForStorage)(log)
		}
		vs.Add(&v0, vserial.VERSION_0)
	}
	if r.PoolId != nil || r.ShareId != nil {
		v1 := &receiptForStorage_Version_1{}
		v1.ShareId = r.ShareId
		v1.PoolId = r.PoolId
		vs.Add(&v1, vserial.VERSION_1)
	}
	return rlp.Encode(w, &vs)
}

// DecodeRLP implements rlp.Decoder, and loads both consensus and implementation
// fields of a receipt from an RLP stream.
func (r *ReceiptForStorage) DecodeRLP(s *rlp.Stream) error {
	var v0 receiptForStorage_Version_0
	var v1 receiptForStorage_Version_1
	vs := vserial.NewVSerial()
	vs.Add(&v0, vserial.VERSION_0)
	vs.Add(&v1, vserial.VERSION_1)
	if err := s.Decode(&vs); err != nil {
		return err
	}

	if err := (*Receipt)(r).setStatus(v0.PostStateOrStatus); err != nil {
		return err
	}
	{
		// Assign the consensus fields
		r.CumulativeGasUsed = v0.CumulativeGasUsed
		r.Bloom = v0.Bloom
		r.Logs = make([]*Log, len(v0.Logs))
		for i, log := range v0.Logs {
			r.Logs[i] = (*Log)(log)
		}
		// Assign the implementation fields
		r.TxHash = v0.TxHash
		r.ContractAddress = v0.ContractAddress
		r.GasUsed = v0.GasUsed
	}
	{
		r.PoolId = v1.PoolId
		r.ShareId = v1.ShareId
	}

	return nil
}
