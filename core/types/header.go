package types

import (
	"encoding/binary"
	"math/big"
	"unsafe"

	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/common/hexutil"
	"github.com/dece-cash/go-dece/crypto/sha3"
	"github.com/dece-cash/go-dece/rlp"
)

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() ([]byte, error) {
	return hexutil.Bytes(n[:]).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *BlockNonce) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("BlockNonce", input, n[:])
}

//go:generate gencodec -type Version_0 -field-override headerMarshaling -out gen_header_json.go

// Version_0 represents a block header in the Ethereum blockchain.
type Header struct {
	// Data
	ParentHash  common.Hash    `json:"parentHash"       gencodec:"required"`
	Coinbase    common.Address `json:"miner"            gencodec:"required"`
	Root        common.Hash    `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash    `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash    `json:"receiptsRoot"     gencodec:"required"`
	Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int       `json:"difficulty"       gencodec:"required"`
	Number      *big.Int       `json:"number"           gencodec:"required"`
	GasLimit    uint64         `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64         `json:"gasUsed"          gencodec:"required"`
	Time        *big.Int       `json:"timestamp"        gencodec:"required"`
	Extra       []byte         `json:"extraData"        gencodec:"required"`
	// POW
	MixDigest common.Hash `json:"mixHash"          gencodec:"required"`
	Nonce     BlockNonce  `json:"nonce"            gencodec:"required"`
	// POS
	CurrentVotes []HeaderVote
	ParentVotes  []HeaderVote
}

// field type overrides for gencodec
type headerMarshaling struct {
	Difficulty *hexutil.Big
	Number     *hexutil.Big
	GasLimit   hexutil.Uint64
	GasUsed    hexutil.Uint64
	Time       *hexutil.Big
	Extra      hexutil.Bytes
	Hash       common.Hash `json:"hash"` // adds call to Hash() in MarshalJSON
}

// Hash returns the block hash of the header, which is simply the keccak256 hash of its
// RLP encoding.
func (h *Header) Hash() common.Hash {
	return rlpHash(h)
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashPow() common.Hash {
	return rlpHash([]interface{}{
		h.ParentHash,
		h.Coinbase,
		h.Root,
		h.TxHash,
		h.ReceiptHash,
		h.Bloom,
		h.Difficulty,
		h.Number,
		h.GasLimit,
		h.GasUsed,
		h.Time,
		h.Extra,
	})
}

func (h *Header) HashPos() (ret common.Hash) {
	m := sha3.NewKeccak256()
	m.Write(h.MixDigest[:])
	m.Write(h.Nonce[:])
	hp := m.Sum(nil)
	copy(ret[:], hp)
	return
}

func StakeHash(currentHashPos *common.Hash, parentHashPos *common.Hash, isPool bool) (ret common.Hash) {
	m := sha3.NewKeccak256()
	m.Write(currentHashPos[:])
	m.Write(parentHashPos[:])
	if isPool {
		m.Write([]byte{1})
	}
	sh := m.Sum(nil)
	copy(ret[:], sh)
	return
}

func (h *Header) ActualDifficulty() *big.Int {
	return h.Difficulty
}

// Size returns the approximate memory used by all internal contents. It is used
// to approximate and limit the memory consumption of various caches.
func (h *Header) Size() common.StorageSize {
	return common.StorageSize(unsafe.Sizeof(*h)) + common.StorageSize(len(h.Extra)+(h.Difficulty.BitLen()+h.Number.BitLen()+h.Time.BitLen())/8)
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// CopyHeader creates a deep copy of a block header to prevent side effects from
// modifying a header variable.
func CopyHeader(h *Header) *Header {
	cpy := *h
	if cpy.Time = new(big.Int); h.Time != nil {
		cpy.Time.Set(h.Time)
	}
	if cpy.Difficulty = new(big.Int); h.Difficulty != nil {
		cpy.Difficulty.Set(h.Difficulty)
	}
	if cpy.Number = new(big.Int); h.Number != nil {
		cpy.Number.Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	if len(h.CurrentVotes) > 0 {
		cpy.CurrentVotes = append([]HeaderVote{}, h.CurrentVotes...)
	}
	if len(h.ParentVotes) > 0 {
		cpy.ParentVotes = append([]HeaderVote{}, h.ParentVotes...)
	}
	return &cpy
}

// // DecodeRLP decodes the Ethereum
// func (b *Header) DecodeRLP(s *rlp.Stream) error {
// 	if e := s.Decode(b); e != nil {
// 		return e
// 	}
// 	return nil
// }
//
// // EncodeRLP serializes b into the Ethereum RLP block format.
// func (b *Header) EncodeRLP(w io.Writer) error {
// 	return rlp.Encode(w, b)
// }
