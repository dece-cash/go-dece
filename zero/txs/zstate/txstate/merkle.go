package txstate

import (
	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/crypto"
	"github.com/dece-cash/go-dece/zero/txs/zstate/merkle"
)

var CzeroAddress = c_type.NewPKrByBytes(crypto.Keccak512(nil))
var CzeroMerkleParam = merkle.NewParam(&CzeroAddress, c_superzk.Czero_combine)

var SzkAddress = c_type.NewPKrByBytes(crypto.Keccak256([]byte("$SuperZK$MerkleTree")))
var SzkMerkleParam = merkle.NewParam(&SzkAddress, c_superzk.Combine)
