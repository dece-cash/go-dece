// Copyright 2015 The go-ethereum Authors
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

// Package ethapi implements the general Ethereum API functions.
package ethapi

import (
	"context"
	"math/big"

	"github.com/dece-cash/go-dece/zero/txtool/prepare"

	"github.com/dece-cash/go-dece/zero/txtool"

	"github.com/dece-cash/go-dece/zero/wallet/exchange"

	"github.com/dece-cash/go-dece/czero/c_type"

	"github.com/dece-cash/go-dece/miner"

	"github.com/dece-cash/go-dece/accounts"
	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/consensus"
	"github.com/dece-cash/go-dece/core"
	"github.com/dece-cash/go-dece/core/state"
	"github.com/dece-cash/go-dece/core/types"
	"github.com/dece-cash/go-dece/core/vm"
	"github.com/dece-cash/go-dece/event"
	"github.com/dece-cash/go-dece/params"
	"github.com/dece-cash/go-dece/rpc"
	"github.com/dece-cash/go-dece/dece/downloader"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/wallet/light"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General Ethereum API
	Downloader() *downloader.Downloader
	ProtocolVersion() int
	PeerCount() uint
	SuggestPrice(ctx context.Context) (*big.Int, error)
	ChainDb() decedb.Database
	EventMux() *event.TypeMux
	AccountManager() *accounts.Manager

	// BlockChain API
	SetHead(number uint64)
	HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error)
	BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error)
	StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error)
	GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error)
	GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
	GetTd(blockHash common.Hash) *big.Int
	GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error)
	SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription
	SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription
	SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription

	// TxPool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	//GetPoolNonce(ctx context.Context, addr common.Data) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (types.Transactions, types.Transactions)
	SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription

	ChainConfig() *params.ChainConfig
	CurrentBlock() *types.Block
	GetEngin() consensus.Engine
	GetMiner() *miner.Miner

	GetBlocksInfo(start uint64, count uint64) ([]txtool.Block, error)
	GetAnchor(roots []c_type.Uint256) ([]txtool.Witness, error)
	CommitTx(tx *txtool.GTx) error

	GetPkNumber(pk c_type.Uint512) (number uint64, e error)
	GetPkr(pk *c_type.Uint512, index *c_type.Uint256) (c_type.PKr, error)
	GetBalances(pk c_type.Uint512) (balances map[string]*big.Int, tickets map[string][]*common.Hash)
	GenTx(param prepare.PreTxParam) (*txtool.GTxParam, error)
	GetRecordsByPk(pk *c_type.Uint512, begin, end uint64) (records []exchange.Utxo, err error)
	GetRecordsByPkr(pkr c_type.PKr, begin, end uint64) (records []exchange.Utxo, err error)
	GetLockedBalances(pk c_type.Uint512) (balances map[string]*big.Int)
	GetMaxAvailable(pk c_type.Uint512, currency string) (amount *big.Int)
	GetRecordsByTxHash(txHash c_type.Uint256) (records []exchange.Utxo, err error)

	//Light node api
	GetOutByPKr(pkrs []c_type.PKr, start, end uint64) (br light.BlockOutResp, e error)
	CheckNil(Nils []c_type.Uint256) (nilResps []light.NilValue, e error)
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	return []rpc.API{
		{
			Namespace: "proof",
			Version:   "1.0",
			Service:   NewProofServiceApi(),
			Public:    true,
		},
		{
			Namespace: "stake",
			Version:   "1.0",
			Service:   NewPublicStakeApI(apiBackend, nonceLock),
			Public:    true,
		},
		{
			Namespace: "dece",
			Version:   "1.0",
			Service:   &PublicAbiAPI{},
			Public:    true,
		},
		{
			Namespace: "light",
			Version:   "1.0",
			Service:   &PublicLightNodeApi{apiBackend},
			Public:    true,
		},
		{
			Namespace: "ssi",
			Version:   "1.0",
			Service:   &PublicSSIAPI{apiBackend},
			Public:    true,
		},
		{
			Namespace: "local",
			Version:   "1.0",
			Service:   &PublicLocalAPI{},
			Public:    true,
		},
		{
			Namespace: "flight",
			Version:   "1.0",
			Service:   &PublicFlightAPI{&PublicExchangeAPI{apiBackend}},
			Public:    true,
		},
		{
			Namespace: "exchange",
			Version:   "1.0",
			Service:   &PublicExchangeAPI{apiBackend},
			Public:    true,
		},
		{
			Namespace: "dece",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "dece",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "dece",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "dece",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		},
	}
}
