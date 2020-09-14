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

// Package dece implements the Dece protocol.
package dece

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/dece-cash/go-dece/zero/wallet/stakeservice"

	"github.com/dece-cash/go-dece/common/address"

	"github.com/dece-cash/go-dece/voter"
	"github.com/dece-cash/go-dece/zero/txtool"
	"github.com/dece-cash/go-dece/zero/zconfig"

	"github.com/dece-cash/go-dece/internal/ethapi"
	"github.com/dece-cash/go-dece/zero/wallet/exchange"

	"github.com/dece-cash/go-dece/accounts"
	"github.com/dece-cash/go-dece/common/hexutil"
	"github.com/dece-cash/go-dece/consensus"
	"github.com/dece-cash/go-dece/consensus/ethash"
	"github.com/dece-cash/go-dece/core"
	"github.com/dece-cash/go-dece/core/bloombits"
	"github.com/dece-cash/go-dece/core/rawdb"
	"github.com/dece-cash/go-dece/core/types"
	"github.com/dece-cash/go-dece/core/vm"
	"github.com/dece-cash/go-dece/event"
	"github.com/dece-cash/go-dece/log"
	"github.com/dece-cash/go-dece/miner"
	"github.com/dece-cash/go-dece/node"
	"github.com/dece-cash/go-dece/p2p"
	"github.com/dece-cash/go-dece/params"
	"github.com/dece-cash/go-dece/rlp"
	"github.com/dece-cash/go-dece/rpc"
	"github.com/dece-cash/go-dece/dece/downloader"
	"github.com/dece-cash/go-dece/dece/filters"
	"github.com/dece-cash/go-dece/dece/gasprice"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/wallet/light"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Dece implements the Dece full node service.
type Dece struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the Dece

	// Handlers
	txPool          *core.TxPool
	voter           *voter.Voter
	blockchain      *core.BlockChain
	exchange        *exchange.Exchange
	lightNode       *light.LightNode
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb decedb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	APIBackend *DeceAPIBackend

	miner    *miner.Miner
	gasPrice *big.Int
	decebase accounts.Account

	networkID     uint64
	netRPCService *ethapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (s.g. gas price and decebase)
}

func (s *Dece) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

var DeceInstance *Dece

// New creates a new Dece object (including the
// initialisation of the common Dece object)
func New(ctx *node.ServiceContext, config *Config) (*Dece, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run dece.Dece in light sync mode, use les.LightEthereum")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	dece := &Dece{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Ethash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkID:      config.NetworkId,
		gasPrice:       config.GasPrice,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising Dece protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gece upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	dece.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, dece.chainConfig, dece.engine, vmConfig, dece.accountManager)

	txtool.Ref_inst.SetBC(&core.State1BlockChain{dece.blockchain})

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		dece.blockchain.SetHead(compat.RewindTo, core.DelFn)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	dece.bloomIndexer.Start(dece.blockchain)

	// if config.TxPool.Journal != "" {
	//	config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	// }

	config.TxPool.StartLight = config.StartLight

	dece.txPool = core.NewTxPool(config.TxPool, dece.chainConfig, dece.blockchain)

	dece.voter = voter.NewVoter(dece.chainConfig, dece.blockchain, dece)

	if dece.protocolManager, err = NewProtocolManager(dece.chainConfig, config.SyncMode, config.NetworkId, dece.eventMux, dece.voter, dece.txPool, dece.engine, dece.blockchain, chainDb); err != nil {
		return nil, err
	}
	dece.miner = miner.New(dece, dece.chainConfig, dece.EventMux(), dece.voter, dece.engine)
	dece.miner.SetExtra(makeExtraData(config.ExtraData))

	dece.APIBackend = &DeceAPIBackend{dece, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	dece.APIBackend.gpo = gasprice.NewOracle(dece.APIBackend, gpoParams)

	ethapi.Backend_Instance = dece.APIBackend

	// init exchange
	if config.StartExchange {
		dece.exchange = exchange.NewExchange(zconfig.Exchange_dir(), dece.txPool, dece.accountManager, config.AutoMerge)
	}

	if config.StartStake {
		stakeservice.NewStakeService(zconfig.Stake_dir(), dece.blockchain, dece.accountManager)
	}

	// init light
	if config.StartLight {
		dece.lightNode = light.NewLightNode(zconfig.Light_dir(), dece.txPool, dece.blockchain.GetDB())
	}

	// if config.Proof != nil {
	// 	if config.Proof.PKr == (c_type.PKr{}) {
	// 		wallets := dece.accountManager.Wallets()
	// 		if len(wallets) == 0 {
	// 			// panic("init proofService error")
	// 		}
	//
	// 		account := wallets[0].Accounts()
	// 		config.Proof.PKr = superzk.Pk2PKr(account[0].Address.ToUint512(), &c_type.Uint256{1})
	// 	}
	// 	proofservice.NewProofService("", dece.APIBackend, config.Proof);
	// }

	DeceInstance = dece
	return dece, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gece",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (decedb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*decedb.LDBDatabase); ok {
		db.Meter("dece/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Dece service
func CreateConsensusEngine(ctx *node.ServiceContext, config *ethash.Config, chainConfig *params.ChainConfig, db decedb.Database) consensus.Engine { // If proof-of-authority is requested, set it up
	// Otherwise assume proof-of-work
	switch config.PowMode {
	case ethash.ModeFake:
		log.Warn("Ethash used in fake mode")
		return ethash.NewFaker()
	case ethash.ModeTest:
		log.Warn("Ethash used in test mode")
		return ethash.NewTester()
	case ethash.ModeShared:
		log.Warn("Ethash used in shared mode")
		return ethash.NewShared()
	default:
		engine := ethash.New(ethash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs return the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Dece) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "dece",
			Version:   "1.0",
			Service:   NewPublicSeroAPI(s),
			Public:    true,
		}, {
			Namespace: "dece",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "dece",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "dece",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Dece) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Dece) Serobase() (eb accounts.Account, err error) {
	s.lock.RLock()
	decebase := s.decebase
	s.lock.RUnlock()

	if decebase != (accounts.Account{}) {
		return decebase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			decebase := accounts[0]

			s.lock.Lock()
			s.decebase = decebase
			s.lock.Unlock()

			log.Info("Serobase automatically configured", "address", decebase)
			return decebase, nil
		}
	}
	return accounts.Account{}, fmt.Errorf("Serobase must be explicitly specified")
}

// SetSerobase sets the mining reward address.
func (s *Dece) SetSerobase(decebase address.MixBase58Adrress) {
	s.lock.Lock()
	account, _ := s.accountManager.FindAccountByPkr(decebase.ToPkr())
	s.decebase = account
	s.lock.Unlock()

	s.miner.SetSerobase(account)
}

func (s *Dece) StartMining(local bool) error {
	eb, err := s.Serobase()
	if err != nil {
		log.Error("Cannot start mining without decebase", "err", err)
		return fmt.Errorf("decebase missing: %v", err)
	}

	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so none will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Dece) StopMining()         { s.miner.Stop() }
func (s *Dece) IsMining() bool      { return s.miner.Mining() }
func (s *Dece) Miner() *miner.Miner { return s.miner }

func (s *Dece) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Dece) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Dece) TxPool() *core.TxPool               { return s.txPool }
func (s *Dece) Voter() *voter.Voter                { return s.voter }
func (s *Dece) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Dece) Engine() consensus.Engine           { return s.engine }
func (s *Dece) ChainDb() decedb.Database           { return s.chainDb }
func (s *Dece) IsListening() bool                  { return true } // Always listening
func (s *Dece) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Dece) NetVersion() uint64                 { return s.networkID }
func (s *Dece) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Dece) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Dece protocol implementation.
func (s *Dece) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Dece protocol.
func (s *Dece) Stop() error {
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
