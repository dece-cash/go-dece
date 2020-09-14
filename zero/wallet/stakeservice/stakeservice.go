package stakeservice

import (
	"github.com/dece-cash/go-dece/rlp"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/robfig/cron"
	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/common/math"
	"github.com/dece-cash/go-dece/zero/utils"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/czero/deceparam"
	"github.com/dece-cash/go-dece/czero/superzk"
	"github.com/dece-cash/go-dece/accounts"
	"github.com/dece-cash/go-dece/core"
	"github.com/dece-cash/go-dece/event"
	"github.com/dece-cash/go-dece/log"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/stake"
)

type Account struct {
	pk      *c_type.Uint512
	tk      *c_type.Tk
	version int
}

type StakeService struct {
	bc             *core.BlockChain
	accountManager *accounts.Manager
	db             *decedb.LDBDatabase

	accounts sync.Map
	numbers  sync.Map

	feed    event.Feed
	updater event.Subscription        // Wallet update subscriptions for all backends
	update  chan accounts.WalletEvent // Subscription sink for backend wallet changes
	quit    chan chan error
	lock    sync.RWMutex
}

var current_StakeService *StakeService

func CurrentStakeService() *StakeService {
	return current_StakeService
}

func NewStakeService(dbpath string, bc *core.BlockChain, accountManager *accounts.Manager) *StakeService {
	update := make(chan accounts.WalletEvent, 1)
	updater := accountManager.Subscribe(update)

	stakeService := &StakeService{
		bc:             bc,
		accountManager: accountManager,
		update:         update,
		updater:        updater,
	}
	current_StakeService = stakeService

	db, err := decedb.NewLDBDatabase(dbpath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	stakeService.db = db

	stakeService.numbers = sync.Map{}
	stakeService.accounts = sync.Map{}
	for _, w := range accountManager.Wallets() {
		stakeService.initWallet(w)
	}

	AddJob("0/10 * * * * ?", stakeService.stakeIndex)
	go stakeService.updateAccount()
	return stakeService
}

func (self *StakeService) StakePools() (pools []*stake.StakePool) {
	iterator := self.db.NewIteratorWithPrefix(poolPrefix)
	for iterator.Next() {

		value := iterator.Value()
		pool := stake.StakePoolDB.GetObject(self.bc.GetDB(), value, &stake.StakePool{})
		pools = append(pools, pool.(*stake.StakePool))
	}
	return
}

func (self *StakeService) Shares() (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(sharePrefix)
	for iterator.Next() {
		value := iterator.Value()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), value, &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) SharesInfoByPKr(pkr c_type.PKr) *SharesInfo {
	hash, err := self.db.Get(pkrInfoKey(pkr))
	if err != nil {
		return nil
	}
	item := &SharesInfo{}
	if e := rlp.DecodeBytes(hash, item); e != nil {
		return nil
	}
	return item
}

func (self *StakeService) SharesInfoByPK(pk c_type.Uint512) *SharesInfo {
	hash, err := self.db.Get(pkInfoKey(pk))
	if err != nil {
		return nil
	}
	item := &SharesInfo{}
	if e := rlp.DecodeBytes(hash, item); e != nil {
		return nil
	}
	return item
}

func (self *StakeService) SharesById(id common.Hash) *stake.Share {
	hash, err := self.db.Get(sharekey(id[:]))
	if err != nil {
		return nil
	}
	return self.getShareByHash(hash)
}

func (self *StakeService) getShareByHash(hash []byte) *stake.Share {
	ret := stake.ShareDB.GetObject(self.bc.GetDB(), hash, &stake.Share{})
	if ret == nil {
		return nil
	}
	return ret.(*stake.Share)
}

func (self *StakeService) SharesByPk(pk c_type.Uint512) (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(pk[:])
	for iterator.Next() {
		value := iterator.Value()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), value, &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) SharesInfoByPkr(pkr c_type.PKr) (sharesInfo *SharesInfo) {
	if date, err := self.db.Get(pkrInfoKey(pkr)); err != nil {
		return
	} else {
		sharesInfo = &SharesInfo{}
		if e := rlp.DecodeBytes(date, sharesInfo); e != nil {
			return nil
		}
		return sharesInfo
	}
	return
}

func (self *StakeService) SharesInfoByPk(pk c_type.Uint512) (sharesInfo *SharesInfo) {
	if date, err := self.db.Get(pkInfoKey(pk)); err != nil {
		return
	} else {
		sharesInfo = &SharesInfo{}
		if e := rlp.DecodeBytes(date, sharesInfo); e != nil {
			return nil
		}
		return sharesInfo
	}
	return
}

func (self *StakeService) SharesByPkr(pkr c_type.PKr) (shares []*stake.Share) {
	iterator := self.db.NewIteratorWithPrefix(pkr[:])
	for iterator.Next() {
		value := iterator.Value()
		share := stake.ShareDB.GetObject(self.bc.GetDB(), value, &stake.Share{})
		shares = append(shares, share.(*stake.Share))
	}
	return
}

func (self *StakeService) GetBlockRecords(blockNumber uint64) (shares []*stake.Share, pools []*stake.StakePool) {
	header := self.bc.GetHeaderByNumber(blockNumber)
	return stake.GetBlockRecords(self.bc.GetDB(), header.Hash(), blockNumber)
}

type SharesInfo struct {
	Total       uint32
	Remaining   uint32
	Missed      uint32
	Expired     uint32
	ShareIds    []common.Hash
	Profit      *big.Int `rlp:"nil"`
	TotalAmount *big.Int `rlp:"nil"`
}

func (self *StakeService) getShare(id common.Hash, cache map[common.Hash]*stake.Share) *stake.Share {
	if val, ok := cache[id]; ok {
		return val
	} else {
		return self.SharesById(id)
	}
}

func (self *StakeService) getStakeInfoByPKr(pkr c_type.PKr, cache map[c_type.PKr]*SharesInfo) *SharesInfo {
	if val, ok := cache[pkr]; ok {
		return val
	} else {
		info := self.SharesInfoByPKr(pkr)
		if info != nil {
			cache[pkr] = info
		}
		return info
	}
}

func (self *StakeService) getStakeInfoByPK(pk c_type.Uint512, cache map[c_type.Uint512]*SharesInfo) *SharesInfo {
	if val, ok := cache[pk]; ok {
		return val
	} else {
		info := self.SharesInfoByPK(pk)
		if info != nil {
			cache[pk] = info
		}
		return info
	}
}

func (self *StakeService) stakeIndex() {
	start := uint64(math.MaxUint64)
	self.numbers.Range(func(key, value interface{}) bool {
		num := value.(uint64)
		if start > num {
			start = num
		}
		return true
	})
	if start == uint64(math.MaxUint64) {
		return
	}
	if !deceparam.Is_Dev() && start < 1300000 {
		start = 1300000
	}

	header := self.bc.CurrentHeader()
	sharesCount := 0
	poolsCount := 0
	batch := self.db.NewBatch()
	blocNumber := start
	sharesCache := map[common.Hash]*stake.Share{}
	pkrStakeInfoCache := map[c_type.PKr]*SharesInfo{}
	pkStakeInfoCache := map[c_type.Uint512]*SharesInfo{}
	for blocNumber+deceparam.DefaultConfirmedBlock() <= header.Number.Uint64() {
		shares, pools := self.GetBlockRecords(blocNumber)
		for _, share := range shares {
			// batch.Put(sharekey(share.Id()), share.State())
			// batch.Put(pkrShareKey(share.PKr, share.Id()), share.State())

			self.indexStakeInfoByPKr(share.PKr, pkrStakeInfoCache, share, sharesCache, blocNumber, batch)
			if accountKey, ok := self.ownPkr(share.PKr); ok {
				// batch.Put(pkShareKey(accountKey, share.Id()), share.State())
				self.indexStakeInfoByPK(*accountKey, pkStakeInfoCache, share, sharesCache, blocNumber, batch)
			}
			sharesCache[common.BytesToHash(share.Id())] = share
		}

		for _, pool := range pools {
			batch.Put(poolKey(pool.Id()), pool.State())
		}
		sharesCount += len(shares)
		poolsCount += len(pools)
		blocNumber++
		if blocNumber-start >= 10000 {
			break
		}
	}

	for _, share := range sharesCache {
		hash := share.State()
		id := share.Id()
		batch.Put(sharekey(id), hash)
		batch.Put(pkrShareKey(share.PKr, id), hash)

		if accountKey, ok := self.ownPkr(share.PKr); ok {
			batch.Put(pkShareKey(accountKey, id), hash)
		}
	}

	for key, val := range pkrStakeInfoCache {
		if b, err := rlp.EncodeToBytes(val); err != nil {
			panic(err)
		} else {
			if err := batch.Put(pkrInfoKey(key), b); err != nil {
				panic(err)
			}
		}
	}

	for key, val := range pkStakeInfoCache {
		if b, err := rlp.EncodeToBytes(&val); err != nil {
			panic(err)
		} else {
			if err := batch.Put(pkInfoKey(key), b); err != nil {
				panic(err)
			}
		}
	}

	if blocNumber == start {
		return
	}

	self.numbers.Range(func(key, value interface{}) bool {
		pk := key.(c_type.Uint512)
		batch.Put(numKey(pk), utils.EncodeNumber(blocNumber))
		return true
	})
	err := batch.Write()
	if err == nil {
		self.numbers.Range(func(key, value interface{}) bool {
			pk := key.(c_type.Uint512)
			self.numbers.Store(pk, blocNumber)
			return true
		})
		log.Info("StakeIndex", "blockNumber", blocNumber, "sharesCount", sharesCount, "poolsCount", poolsCount)
	}
}

func (self *StakeService) indexStakeInfoByPKr(pkr c_type.PKr, stakeInfoCache map[c_type.PKr]*SharesInfo, share *stake.Share, sharesCache map[common.Hash]*stake.Share, blocNumber uint64, batch decedb.Batch) {
	sharesInfo := self.getStakeInfoByPKr(pkr, stakeInfoCache)
	if sharesInfo == nil {
		sharesInfo = &SharesInfo{
			Profit:      new(big.Int),
			TotalAmount: new(big.Int),
		}
		stakeInfoCache[pkr] = sharesInfo
	}

	self.updateSharesInfo(share, sharesCache, sharesInfo, blocNumber)
}

func (self *StakeService) indexStakeInfoByPK(pk c_type.Uint512, stakeInfoCache map[c_type.Uint512]*SharesInfo, share *stake.Share, sharesCache map[common.Hash]*stake.Share, blocNumber uint64, batch decedb.Batch) {
	sharesInfo := self.getStakeInfoByPK(pk, stakeInfoCache)
	if sharesInfo == nil {
		sharesInfo = &SharesInfo{
			Profit:      new(big.Int),
			TotalAmount: new(big.Int),
		}
		stakeInfoCache[pk] = sharesInfo
	}

	self.updateSharesInfo(share, sharesCache, sharesInfo, blocNumber)
}

func (self *StakeService) updateSharesInfo(share *stake.Share, sharesCache map[common.Hash]*stake.Share, sharesInfo *SharesInfo, blocNumber uint64) {
	id := common.BytesToHash(share.Id())

	oldShare := self.getShare(id, sharesCache)
	// sharesCache[id] = share
	if oldShare != nil {
		if share.WillVoteNum > oldShare.WillVoteNum {
			sharesInfo.Missed += (share.WillVoteNum - oldShare.WillVoteNum)
		} else if share.WillVoteNum < oldShare.WillVoteNum {
			if sharesInfo.Missed >= (oldShare.WillVoteNum - share.WillVoteNum) {
				sharesInfo.Missed -= (oldShare.WillVoteNum - share.WillVoteNum)
			}
		}

		if oldShare.Status == stake.STATUS_VALID {
			if sharesInfo.Remaining >= (oldShare.Num - share.Num) {
				sharesInfo.Remaining -= (oldShare.Num - share.Num)
			}
		}

		if oldShare.Status == stake.STATUS_VALID && share.Status == stake.STATUS_OUTOFDATE {
			if sharesInfo.Remaining >= share.Num {
				sharesInfo.Remaining -= share.Num
			}
			sharesInfo.Expired += share.Num
		}
		sharesInfo.Profit = big.NewInt(0).Add(sharesInfo.Profit, new(big.Int).Sub(share.Profit, oldShare.Profit))
	} else {
		sharesInfo.ShareIds = append(sharesInfo.ShareIds, id)
		sharesInfo.Total += share.InitNum

		if share.Status == stake.STATUS_VALID {
			sharesInfo.Remaining += share.Num
		} else {
			sharesInfo.Expired += share.Num
		}

		sharesInfo.Missed += share.WillVoteNum
		sharesInfo.Profit = new(big.Int).Add(sharesInfo.Profit, share.Profit)
		sharesInfo.TotalAmount = new(big.Int).Add(sharesInfo.TotalAmount, new(big.Int).Mul(big.NewInt(int64(share.InitNum)), share.Value))
	}

	if share.LastPayTime == blocNumber {
		if oldShare != nil && oldShare.LastPayTime != 0 {
			header := self.bc.GetBlockByNumber(oldShare.LastPayTime)
			snapshot := stake.GetShareByBlockNumber(self.bc.GetDB(), id, header.Hash(), header.NumberU64())
			if snapshot != nil {
				mul := new(big.Int).Mul(big.NewInt(int64(
					(snapshot.Num+snapshot.WillVoteNum)-(share.Num+share.WillVoteNum))),
					share.Value)
				if sharesInfo.TotalAmount.Cmp(mul) >= 0 {
					sharesInfo.TotalAmount = new(big.Int).Sub(sharesInfo.TotalAmount, mul)
				}
			}
		} else {
			mul := new(big.Int).Mul(big.NewInt(int64(
				share.InitNum-share.Num-share.WillVoteNum)),
				share.Value)
			if sharesInfo.TotalAmount.Cmp(mul) >= 0 {
				sharesInfo.TotalAmount = new(big.Int).Sub(sharesInfo.TotalAmount, mul)
			}
		}

		if share.Status == stake.STATUS_OUTOFDATE {
			mul := new(big.Int).Mul(big.NewInt(int64(
				share.Num)),
				share.Value)
			if sharesInfo.TotalAmount.Cmp(mul) >= 0 {
				sharesInfo.TotalAmount = new(big.Int).Sub(sharesInfo.TotalAmount, mul)
			}
		}
		if share.Status == stake.STATUS_FINISHED {
			mul := new(big.Int).Mul(big.NewInt(int64(
				share.WillVoteNum)),
				share.Value)
			if sharesInfo.TotalAmount.Cmp(mul) >= 0 {
				sharesInfo.TotalAmount = new(big.Int).Sub(sharesInfo.TotalAmount, mul)
			}
		}
	}

	if stake.STATUS_FINISHED == share.Status {
		for i, each := range sharesInfo.ShareIds {
			if each == id {
				sharesInfo.ShareIds = append(sharesInfo.ShareIds[:i], sharesInfo.ShareIds[i+1:]...)
				break
			}
		}
	}
}

func (self *StakeService) ownPkr(pkr c_type.PKr) (pk *c_type.Uint512, ok bool) {
	var account *Account
	self.accounts.Range(func(key, value interface{}) bool {
		a := value.(*Account)
		if superzk.IsMyPKr(a.tk, &pkr) {
			account = a
			return false
		}
		return true
	})
	if account != nil {
		return account.pk.NewRef(), true
	}
	return
}

func (self *StakeService) updateAccount() {
	// Close all subscriptions when the manager terminates
	defer func() {
		self.lock.Lock()
		self.updater.Unsubscribe()
		self.updater = nil
		self.lock.Unlock()
	}()

	// Loop until termination
	for {
		select {
		case event := <-self.update:
			// Wallet event arrived, update local cache
			self.lock.Lock()
			switch event.Kind {
			case accounts.WalletArrived:
				self.initWallet(event.Wallet)
			case accounts.WalletDropped:
				address := event.Wallet.Accounts()[0].GetPk()
				self.numbers.Delete(address)
			}
			self.lock.Unlock()

		case errc := <-self.quit:
			// Manager terminating, return
			errc <- nil
			return
		}
	}
}

func (self *StakeService) initWallet(w accounts.Wallet) {
	if _, ok := self.accounts.Load(w.Accounts()[0].GetPk()); !ok {
		account := Account{}
		account.pk = w.Accounts()[0].GetPk().NewRef()
		account.tk = w.Accounts()[0].Tk.ToTk().NewRef()
		account.version = w.Accounts()[0].Version
		self.accounts.Store(*account.pk, &account)

		var num uint64
		if num = self.starNum(account.pk); num < w.Accounts()[0].At {
			num = w.Accounts()[0].At
		}
		self.numbers.Store(*account.pk, num)
		log.Info("Add PK", "pk", w.Accounts()[0].Address, "At", num)
	}
}

func (self *StakeService) starNum(pk *c_type.Uint512) uint64 {
	value, err := self.db.Get(numKey(*pk))
	if err != nil {
		return 0
	}
	return utils.DecodeNumber(value)
}

var (
	numPrefix     = []byte("NUM")
	sharePrefix   = []byte("SHARE")
	poolPrefix    = []byte("POOL")
	pkrInfoPrefix = []byte("PKRINFO")
	pkInfoPrefix  = []byte("PKINFO")
)

func pkShareKey(pk *c_type.Uint512, key []byte) []byte {
	return append(pk[:], key[:]...)
}

func pkrShareKey(pk c_type.PKr, key []byte) []byte {
	return append(pk[:], key[:]...)
}

func sharekey(key []byte) []byte {
	return append(sharePrefix, key[:]...)
}

func poolKey(key []byte) []byte {
	return append(poolPrefix, key[:]...)
}

func numKey(pk c_type.Uint512) []byte {
	return append(numPrefix, pk[:]...)
}

func pkrInfoKey(pkr c_type.PKr) []byte {
	return append(pkrInfoPrefix, pkr[:]...)
}

func pkInfoKey(pk c_type.Uint512) []byte {
	return append(pkInfoPrefix, pk[:]...)
}

func AddJob(spec string, run RunFunc) *cron.Cron {
	c := cron.New()
	c.AddJob(spec, &RunJob{run: run})
	c.Start()
	return c
}

type (
	RunFunc func()
)

type RunJob struct {
	runing int32
	run    RunFunc
}

func (r *RunJob) Run() {
	x := atomic.LoadInt32(&r.runing)
	if x == 1 {
		return
	}

	atomic.StoreInt32(&r.runing, 1)
	defer func() {
		atomic.StoreInt32(&r.runing, 0)
	}()

	r.run()
}
