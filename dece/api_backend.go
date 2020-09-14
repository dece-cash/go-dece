// copyright 2018 The go-ethereum Authors
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

package dece

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/dece-cash/go-dece/zero/txtool/flight"

	"github.com/dece-cash/go-dece/zero/txtool"
	"github.com/dece-cash/go-dece/zero/txtool/prepare"

	"github.com/dece-cash/go-dece/zero/wallet/exchange"

	"github.com/dece-cash/go-dece/log"

	"github.com/dece-cash/go-dece/czero/c_type"

	"github.com/dece-cash/go-dece/consensus"
	"github.com/dece-cash/go-dece/miner"

	"github.com/dece-cash/go-dece/accounts"
	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/core"
	"github.com/dece-cash/go-dece/core/bloombits"
	"github.com/dece-cash/go-dece/core/rawdb"
	"github.com/dece-cash/go-dece/core/state"
	"github.com/dece-cash/go-dece/core/types"
	"github.com/dece-cash/go-dece/core/vm"
	"github.com/dece-cash/go-dece/event"
	"github.com/dece-cash/go-dece/params"
	"github.com/dece-cash/go-dece/rpc"
	"github.com/dece-cash/go-dece/dece/downloader"
	"github.com/dece-cash/go-dece/dece/gasprice"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/wallet/light"
)

// DeceAPIBackend implements ethapi.Backend for full nodes
type DeceAPIBackend struct {
	dece *Dece
	gpo  *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *DeceAPIBackend) ChainConfig() *params.ChainConfig {
	return b.dece.chainConfig
}

func (b *DeceAPIBackend) CurrentBlock() *types.Block {
	return b.dece.blockchain.CurrentBlock()
}

func (b *DeceAPIBackend) GetEngin() consensus.Engine {
	return b.dece.engine
}

func (b *DeceAPIBackend) GetMiner() *miner.Miner {
	return b.dece.miner
}

func (b *DeceAPIBackend) SetHead(number uint64) {
	b.dece.protocolManager.downloader.Cancel()
	b.dece.blockchain.SetHead(number, core.DelFn)
}

func (b *DeceAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.dece.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.dece.blockchain.CurrentBlock().Header(), nil
	}
	return b.dece.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *DeceAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.dece.blockchain.GetHeaderByHash(hash), nil
}

func (b *DeceAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.dece.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.dece.blockchain.CurrentBlock(), nil
	}
	return b.dece.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *DeceAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.dece.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.dece.BlockChain().StateAt(header)
	return stateDb, header, err
}

func (b *DeceAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.dece.blockchain.GetBlockByHash(hash), nil
}

func (b *DeceAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.dece.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.dece.chainDb, hash, *number), nil
	}
	return nil, nil
}

func (b *DeceAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.dece.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.dece.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *DeceAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.dece.blockchain.GetTdByHash(blockHash)
}

func (b *DeceAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.dece.BlockChain(), nil)
	return vm.NewEVM(context, state, b.dece.chainConfig, vmCfg), vmError, nil
}

func (b *DeceAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.dece.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *DeceAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.dece.BlockChain().SubscribeChainEvent(ch)
}

func (b *DeceAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.dece.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *DeceAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.dece.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *DeceAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.dece.BlockChain().SubscribeLogsEvent(ch)
}

func (b *DeceAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.dece.txPool.AddLocal(signedTx)
}

func (b *DeceAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.dece.txPool.Pending()
	if err != nil {
		return nil, err
	}

	return pending, nil
}

func (b *DeceAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.dece.txPool.Get(hash)
}

//func (b *DeceAPIBackend) GetPoolNonce(ctx context.Context, addr common.Data) (uint64, error) {
//	return b.dece.txPool.State().GetNonce(addr), nil
//}

func (b *DeceAPIBackend) Stats() (pending int, queued int) {
	return b.dece.txPool.Stats()
}

func (b *DeceAPIBackend) TxPoolContent() (types.Transactions, types.Transactions) {
	return b.dece.TxPool().Content()
}

func (b *DeceAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return b.dece.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *DeceAPIBackend) Downloader() *downloader.Downloader {
	return b.dece.Downloader()
}

func (b *DeceAPIBackend) ProtocolVersion() int {
	return b.dece.EthVersion()
}

func (b *DeceAPIBackend) PeerCount() uint {
	return uint(b.dece.netRPCService.PeerCount())
}

func (b *DeceAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *DeceAPIBackend) ChainDb() decedb.Database {
	return b.dece.ChainDb()
}

func (b *DeceAPIBackend) EventMux() *event.TypeMux {
	return b.dece.EventMux()
}

func (b *DeceAPIBackend) AccountManager() *accounts.Manager {
	return b.dece.AccountManager()
}

func (b *DeceAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.dece.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *DeceAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.dece.bloomRequests)
	}
}

func (b *DeceAPIBackend) GetBlocksInfo(start uint64, count uint64) ([]txtool.Block, error) {
	return flight.SRI_Inst.GetBlocksInfo(start, count)

}
func (b *DeceAPIBackend) GetAnchor(roots []c_type.Uint256) ([]txtool.Witness, error) {
	return flight.SRI_Inst.GetAnchor(roots)

}
func (b *DeceAPIBackend) CommitTx(tx *txtool.GTx) error {

	difference := time.Now().Unix() - b.CurrentBlock().Time().Int64()
	if difference > 10*60 {
		return errors.New("The current chain is too behind")
	}
	gasPrice := big.Int(tx.GasPrice)
	gas := uint64(tx.Gas)
	signedTx := types.NewTxWithGTx(gas, &gasPrice, &tx.Tx)
	log.Info("commitTx", "txhash", signedTx.Hash().String())
	return b.dece.txPool.AddLocal(signedTx)
}

func (b *DeceAPIBackend) GetPkNumber(pk c_type.Uint512) (number uint64, e error) {
	if b.dece.exchange == nil {
		e = errors.New("not start exchange")
		return
	}
	return b.dece.exchange.GetCurrencyNumber(pk), nil
}

func (b *DeceAPIBackend) GetPkr(pk *c_type.Uint512, index *c_type.Uint256) (pkr c_type.PKr, e error) {
	if b.dece.exchange == nil {
		e = errors.New("not start exchange")
		return
	}
	return b.dece.exchange.GetPkr(pk, index)
}

func (b *DeceAPIBackend) GetLockedBalances(pk c_type.Uint512) (balances map[string]*big.Int) {
	if b.dece.exchange == nil {
		return
	}
	return b.dece.exchange.GetLockedBalances(pk)
}

func (b *DeceAPIBackend) GetMaxAvailable(pk c_type.Uint512, currency string) (amount *big.Int) {
	if b.dece.exchange == nil {
		return
	}
	return b.dece.exchange.GetMaxAvailable(pk, currency)
}

func (b *DeceAPIBackend) GetBalances(pk c_type.Uint512) (balances map[string]*big.Int, tickets map[string][]*common.Hash) {
	if b.dece.exchange == nil {
		return
	}
	return b.dece.exchange.GetBalances(pk)
}

func (b *DeceAPIBackend) GenTx(param prepare.PreTxParam) (txParam *txtool.GTxParam, e error) {
	if b.dece.exchange == nil {
		e = errors.New("not start exchange")
		return
	}
	return b.dece.exchange.GenTx(param)
}

func (b *DeceAPIBackend) GetRecordsByPkr(pkr c_type.PKr, begin, end uint64) (records []exchange.Utxo, err error) {
	if b.dece.exchange == nil {
		err = errors.New("not start exchange")
		return
	}
	return b.dece.exchange.GetRecordsByPkr(pkr, begin, end)
}

func (b *DeceAPIBackend) GetRecordsByPk(pk *c_type.Uint512, begin, end uint64) (records []exchange.Utxo, err error) {
	if b.dece.exchange == nil {
		err = errors.New("not start exchange")
		return
	}
	return b.dece.exchange.GetRecordsByPk(pk, begin, end)
}

func (b *DeceAPIBackend) GetRecordsByTxHash(txHash c_type.Uint256) (records []exchange.Utxo, err error) {
	if b.dece.exchange == nil {
		err = errors.New("not start exchange")
		return
	}
	return b.dece.exchange.GetRecordsByTxHash(txHash)
}

func (b *DeceAPIBackend) GetOutByPKr(pkrs []c_type.PKr, start, end uint64) (br light.BlockOutResp, e error) {
	if b.dece.lightNode == nil {
		e = errors.New("not start light")
		return
	}
	return b.dece.lightNode.GetOutsByPKr(pkrs, start, end)
}

func (b *DeceAPIBackend) CheckNil(Nils []c_type.Uint256) (nilResps []light.NilValue, e error) {
	if b.dece.lightNode == nil {
		e = errors.New("not start light")
		return
	}
	return b.dece.lightNode.CheckNil(Nils)
}
