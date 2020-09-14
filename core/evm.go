// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"
	"strings"

	"github.com/dece-cash/go-dece/zero/txs/assets"

	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/consensus"
	"github.com/dece-cash/go-dece/core/types"
	"github.com/dece-cash/go-dece/core/vm"
)

// ChainContext supports retrieving headers and consensus parameters from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	// Engine retrieves the chain's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(common.Hash, uint64) *types.Header
}

// NewEVMContext creates a new context for use in the EVM.
func NewEVMContext(msg Message, header *types.Header, chain ChainContext, author *common.Address) vm.Context {
	// If we don't have an explicit author (i.e. not mining), extract from the header
	var beneficiary common.Address
	if author == nil {
		beneficiary, _ = chain.Engine().Author(header) // Ignore error, we're past header validation
	} else {
		beneficiary = *author
	}
	return vm.Context{
		Transfer:    Transfer,
		CanTransfer: CanTransfer,
		GetHash:     GetHashFn(header, chain),
		Origin:      msg.From(),
		Coinbase:    beneficiary,
		BlockNumber: new(big.Int).Set(header.Number),
		Time:        new(big.Int).Set(header.Time),
		Difficulty:  new(big.Int).Set(header.Difficulty),
		GasLimit:    header.GasLimit,
		GasPrice:    new(big.Int).Set(msg.GasPrice()),
		TxHash:      msg.TxHash(),
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(ref *types.Header, chain ChainContext) func(n uint64) common.Hash {
	var cache map[uint64]common.Hash

	return func(n uint64) common.Hash {
		// If there's no hash cache yet, make one
		if cache == nil {
			cache = map[uint64]common.Hash{
				ref.Number.Uint64() - 1: ref.ParentHash,
			}
		}
		// Try to fulfill the request from the cache
		if hash, ok := cache[n]; ok {
			return hash
		}
		// Not cached, iterate the blocks and cache the hashes
		for header := chain.GetHeader(ref.ParentHash, ref.Number.Uint64()-1); header != nil; header = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1) {
			cache[header.Number.Uint64()-1] = header.ParentHash
			if n == header.Number.Uint64()-1 {
				return header.ParentHash
			}
		}
		return common.Hash{}
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, asset *assets.Asset) bool {
	flag := true
	if asset == nil {
		return flag
	}
	if asset.Tkn != nil {
		amount := big.Int(asset.Tkn.Value)
		flag = db.GetBalance(addr, strings.Trim(string(asset.Tkn.Currency[:]), string([]byte{0}))).Cmp(&amount) >= 0
	}
	if flag && asset.Tkt != nil {
		category := strings.Trim(string(asset.Tkt.Category[:]), string([]byte{0}))
		return db.OwnTicket(addr, category, common.BytesToHash(asset.Tkt.Value[:]))
	}
	return flag
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, asset *assets.Asset, txHash common.Hash) (alarm bool) {
	if asset == nil {
		return
	}
	if db.IsContract(sender) {
		if asset.Tkn != nil {
			amount := big.Int(asset.Tkn.Value)
			if amount.Sign() > 0 {
				currency := strings.Trim(string(asset.Tkn.Currency[:]), string([]byte{0}))
				db.SubBalance(sender, currency, &amount)
			}
		}
		if asset.Tkt != nil {
			category := strings.Trim(string(asset.Tkt.Category[:]), string([]byte{0}))
			db.RemoveTicket(sender, category, common.BytesToHash(asset.Tkt.Value[:]))
		}
	}

	if db.IsContract(recipient) {
		if asset.Tkn != nil {
			amount := big.Int(asset.Tkn.Value)
			if amount.Sign() > 0 {
				currency := strings.Trim(string(asset.Tkn.Currency[:]), string([]byte{0}))
				db.AddBalance(recipient, currency, &amount)
			}
		}
		if asset.Tkt != nil {
			category := strings.Trim(string(asset.Tkt.Category[:]), string([]byte{0}))
			db.AddTicket(recipient, category, common.BytesToHash(asset.Tkt.Value[:]))
		}
	} else {
		alarm = db.NextZState().AddTxOutWithCheck(recipient, *asset, txHash)
	}
	return
}
