package exchange

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dece-cash/go-dece/czero/superzk"

	"github.com/dece-cash/go-dece/common/address"

	"github.com/dece-cash/go-dece/zero/txtool"
	"github.com/dece-cash/go-dece/zero/txtool/flight"
	"github.com/dece-cash/go-dece/zero/txtool/prepare"

	"github.com/dece-cash/go-dece/common/hexutil"

	"github.com/robfig/cron"
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/accounts"
	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/core"
	"github.com/dece-cash/go-dece/core/types"
	"github.com/dece-cash/go-dece/event"
	"github.com/dece-cash/go-dece/log"
	"github.com/dece-cash/go-dece/rlp"
	"github.com/dece-cash/go-dece/decedb"
	"github.com/dece-cash/go-dece/zero/txs/assets"
	"github.com/dece-cash/go-dece/zero/utils"
)

type Account struct {
	wallet        accounts.Wallet
	pk            *c_type.Uint512
	tk            *c_type.Tk
	skr           c_type.PKr
	mainPkr       c_type.PKr
	balancePkr    *c_type.PKr
	balances      map[string]*big.Int
	tickets       map[string][]*common.Hash
	utxoNums      map[string]uint64
	isChanged     bool
	nextMergeTime time.Time
	version       int
}

type PkrAccount struct {
	Pkr      c_type.PKr
	balances map[string]*big.Int
	num      uint64
}

type UtxoList []Utxo

func (list UtxoList) Len() int {
	return len(list)
}

func (list UtxoList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list UtxoList) Less(i, j int) bool {
	if list[i].flag == list[j].flag {
		return list[i].Asset.Tkn.Value.ToIntRef().Cmp(list[j].Asset.Tkn.Value.ToIntRef()) < 0
	} else {
		return list[i].flag < list[j].flag
	}
}

func (list UtxoList) Roots() (roots prepare.Utxos) {
	for _, utxo := range list {
		roots = append(roots, prepare.Utxo{utxo.Root, utxo.Asset})
	}
	return
}

type (
	HandleUtxoFunc func(utxo Utxo)
)

type PkKey struct {
	key c_type.Uint512
	Num uint64
}

type PkrKey struct {
	pkr c_type.PKr
	num uint64
}

type FetchJob struct {
	start    uint64
	accounts []Account
}

type Exchange struct {
	db             *decedb.LDBDatabase
	txPool         *core.TxPool
	accountManager *accounts.Manager

	accounts    sync.Map
	pkrAccounts sync.Map

	usedFlag sync.Map
	numbers  sync.Map

	feed    event.Feed
	updater event.Subscription        // Wallet update subscriptions for all backends
	update  chan accounts.WalletEvent // Subscription sink for backend wallet changes
	quit    chan chan error
	lock    sync.RWMutex
}

var current_exchange *Exchange

func CurrentExchange() *Exchange {
	return current_exchange
}

func NewExchange(dbpath string, txPool *core.TxPool, accountManager *accounts.Manager, autoMerge bool) (exchange *Exchange) {

	update := make(chan accounts.WalletEvent, 1)
	updater := accountManager.Subscribe(update)

	exchange = &Exchange{
		txPool:         txPool,
		accountManager: accountManager,
		update:         update,
		updater:        updater,
	}
	current_exchange = exchange

	db, err := decedb.NewLDBDatabase(dbpath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	exchange.db = db

	exchange.numbers = sync.Map{}
	exchange.accounts = sync.Map{}
	for _, w := range accountManager.Wallets() {
		exchange.initWallet(w)
	}

	exchange.pkrAccounts = sync.Map{}
	exchange.usedFlag = sync.Map{}

	AddJob("0/10 * * * * ?", exchange.fetchBlockInfo)

	if autoMerge {
		AddJob("0 0/5 * * * ?", exchange.merge)
	}

	go exchange.updateAccount()
	log.Info("Init NewExchange success")
	return
}

func (self *Exchange) initWallet(w accounts.Wallet) {

	if _, ok := self.accounts.Load(w.Accounts()[0].GetPk()); !ok {
		account := Account{}
		account.wallet = w
		account.pk = w.Accounts()[0].GetPk().NewRef()
		account.tk = w.Accounts()[0].Tk.ToTk().NewRef()
		copy(account.skr[:], account.tk[:])
		account.mainPkr = w.Accounts()[0].GetDefaultPkr(1)
		account.isChanged = true
		account.nextMergeTime = time.Now()
		account.version = w.Accounts()[0].Version
		self.accounts.Store(*account.pk, &account)
		balancePkr := self.getBalancePkr(account.pk)
		if balancePkr != nil {
			if superzk.IsMyPKr(account.tk, balancePkr) {
				account.balancePkr = balancePkr
			}
		}

		if num := self.starNum(account.pk); num > w.Accounts()[0].At {
			self.numbers.Store(*account.pk, num)
		} else {
			self.numbers.Store(*account.pk, w.Accounts()[0].At)
		}

		log.Info("Add PK", "pk", w.Accounts()[0].Address, "At", self.GetCurrencyNumber(*account.pk))
	}
}

func (self *Exchange) starNum(pk *c_type.Uint512) uint64 {
	value, err := self.db.Get(numKey(*pk))
	if err != nil {
		return 0
	}
	return utils.DecodeNumber(value)
}

func (self *Exchange) getBalancePkr(pk *c_type.Uint512) *c_type.PKr {
	value, err := self.db.Get(balancPkrKey(*pk))
	if err != nil {
		return nil
	}
	var pkr c_type.PKr
	err = rlp.DecodeBytes(value, &pkr)
	if err != nil {
		return nil
	}
	return &pkr
}

func (self *Exchange) putBalancePkr(pk *c_type.Uint512, pkr c_type.PKr) error {
	data, err := rlp.EncodeToBytes(&pkr)
	if err != nil {
		return err
	}
	err = self.db.Put(balancPkrKey(*pk), data)
	if err != nil {
		return err
	}
	return nil
}

func (self *Exchange) SetBalancePkr(pk *c_type.Uint512, pkr c_type.PKr) error {

	err := self.putBalancePkr(pk, pkr)
	if err != nil {
		return err
	}
	value, ok := self.accounts.Load(*pk)
	if !ok {
		return errors.New("account not exists")
	}
	account := value.(*Account)
	account.balancePkr = &pkr
	self.accounts.Store(pk, account)
	return nil
}

func (self *Exchange) updateAccount() {
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
				// wallet := event.Wallet
				self.initWallet(event.Wallet)
			case accounts.WalletDropped:
				address := event.Wallet.Accounts()[0].Address
				self.numbers.Delete(address.ToUint512())
			}
			self.lock.Unlock()

		case errc := <-self.quit:
			// Manager terminating, return
			errc <- nil
			return
		}
	}
}

func (self *Exchange) GetUtxoNum(pk c_type.Uint512) map[string]uint64 {
	if account := self.getAccountByPk(pk); account != nil {
		return account.utxoNums
	}
	return map[string]uint64{}
}

func (self *Exchange) GetRootByNil(Nil c_type.Uint256) (root *c_type.Uint256) {
	data, err := self.db.Get(nilToRootKey(Nil))
	if err != nil {
		return
	}
	root = &c_type.Uint256{}
	copy(root[:], data[:])
	return
}

func (self *Exchange) GetCurrencyNumber(pk c_type.Uint512) uint64 {
	value, ok := self.numbers.Load(pk)
	if !ok {
		return 0
	}
	if value.(uint64) == 0 {
		return value.(uint64)
	}
	return value.(uint64) - 1
}

func (self *Exchange) GetPkr(pk *c_type.Uint512, index *c_type.Uint256) (pkr c_type.PKr, err error) {
	if index == nil {
		return pkr, errors.New("index must not be empty")
	}
	if new(big.Int).SetBytes(index[:]).Cmp(big.NewInt(100)) < 0 {
		return pkr, errors.New("index must > 100")
	}
	if value, ok := self.accounts.Load(*pk); !ok {
		return pkr, errors.New("not found Pk")
	} else {
		acc := value.(*Account)
		return acc.wallet.Accounts()[0].GetPkr(index), nil

	}

}

func (self *Exchange) ClearUsedFlagForPK(pk *c_type.Uint512) (count int) {
	if _, ok := self.accounts.Load(*pk); ok {
		prefix := append(pkPrefix, pk[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)

		for iterator.Next() {
			key := iterator.Key()
			var root c_type.Uint256
			copy(root[:], key[98:130])
			if _, flag := self.usedFlag.Load(root); flag {
				self.usedFlag.Delete(root)
				count++
			}
		}
	}
	return
}

func (self *Exchange) ClearUsedFlagForRoot(root c_type.Uint256) (count int) {
	if _, flag := self.usedFlag.Load(root); flag {
		self.usedFlag.Delete(root)
		count++
	}
	return
}

func (self *Exchange) GetLockedBalances(pk c_type.Uint512) (balances map[string]*big.Int) {
	if _, ok := self.accounts.Load(pk); ok {
		prefix := append(pkPrefix, pk[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)
		balances = map[string]*big.Int{}

		for iterator.Next() {
			key := iterator.Key()
			var root c_type.Uint256
			copy(root[:], key[98:130])
			if utxo, err := self.getUtxo(root); err == nil {
				if utxo.Asset.Tkn != nil {
					currency := common.BytesToString(utxo.Asset.Tkn.Currency[:])
					if _, flag := self.usedFlag.Load(utxo.Root); flag {
						if amount, ok := balances[currency]; ok {
							amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
						} else {
							balances[currency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
						}
						currency_locked_key := currency + "_locked"
						if amount, ok := balances[currency_locked_key]; ok {
							amount.Add(amount, big.NewInt(1))
						} else {
							balances[currency_locked_key] = big.NewInt(1)
						}
					}
				}
			}
		}
		return balances
	}
	return
}

func (self *Exchange) GetMaxAvailable(pk c_type.Uint512, currency string) (amount *big.Int) {
	currency = strings.ToUpper(currency)
	prefix := append(pkPrefix, append(pk[:], common.LeftPadBytes([]byte(currency), 32)...)...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	amount = new(big.Int)
	count := 0
	for iterator.Next() {
		key := iterator.Key()
		var root c_type.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if utxo.Ignore {
				continue
			}
			if _, flag := self.usedFlag.Load(utxo.Root); !flag {
				if utxo.Asset.Tkn != nil {
					if utxo.IsZ {
						amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
					} else {
						if count < 2500 {
							amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
							count++
						}
					}
				}
			}
		}
	}
	return
}

func (self *Exchange) IgnorePkrUtxos(pkr c_type.PKr, ignore bool) (utxos []Utxo, e error) {
	account := self.getAccountByPkr(pkr)
	if account == nil {
		e = errors.New("not found PK by pkr")
		return
	}
	pk := account.pk

	prefix := append(pkPrefix, pk[:]...)
	iterator := self.db.NewIteratorWithPrefix(prefix)
	for iterator.Next() {
		key := iterator.Key()
		var root c_type.Uint256
		copy(root[:], key[98:130])
		if utxo, err := self.getUtxo(root); err == nil {
			if utxo.Pkr == pkr {
				utxos = append(utxos, utxo)
			}
		} else {
			e = err
			return
		}
	}

	if len(utxos) > 0 {
		batch := self.db.NewBatch()
		for _, utxo := range utxos {
			utxo.Ignore = ignore
			if bs, err := rlp.EncodeToBytes(&utxo); err == nil {
				if err := batch.Put(rootKey(utxo.Root), bs); err != nil {
					e = err
					return
				}
			} else {
				e = err
				return
			}
		}
		if e = batch.Write(); e != nil {
			return
		}
		account.isChanged = true
	}
	return

}

func (self *Exchange) GetBalances(pk c_type.Uint512) (balances map[string]*big.Int, tickets map[string][]*common.Hash) {
	if value, ok := self.accounts.Load(pk); ok {
		account := value.(*Account)
		if account.isChanged {
			prefix := append(pkPrefix, pk[:]...)
			iterator := self.db.NewIteratorWithPrefix(prefix)
			balances = map[string]*big.Int{}
			tickets = make(map[string][]*common.Hash)
			utxoNums := map[string]uint64{}
			for iterator.Next() {
				key := iterator.Key()
				var root c_type.Uint256
				copy(root[:], key[98:130])
				if utxo, err := self.getUtxo(root); err == nil {
					if utxo.Ignore {
						continue
					}
					if utxo.Asset.Tkn != nil {
						currency := common.BytesToString(utxo.Asset.Tkn.Currency[:])
						if amount, ok := balances[currency]; ok {
							amount.Add(amount, utxo.Asset.Tkn.Value.ToIntRef())
							utxoNums[currency] += 1
						} else {
							balances[currency] = new(big.Int).Set(utxo.Asset.Tkn.Value.ToIntRef())
							utxoNums[currency] = 1
						}
					}
					if utxo.Asset.Tkt != nil {
						category := common.BytesToString(utxo.Asset.Tkt.Category[:])
						ticket := common.BytesToHash(utxo.Asset.Tkt.Value[:])
						if _, ok := tickets[category]; ok {
							tickets[category] = append(tickets[category], &ticket)
						} else {
							tickets[category] = []*common.Hash{&ticket}
						}
					}
				}
			}
			account.balances = balances
			account.tickets = tickets
			account.utxoNums = utxoNums
			account.isChanged = false
		} else {
			return account.balances, account.tickets
		}
	}

	return
}

type BlockInfo struct {
	Num  uint64
	Hash c_type.Uint256
	Ins  []c_type.Uint256
	Outs []Utxo
}

func (self *Exchange) GetBlocksInfo(start, end uint64) (blocks []BlockInfo, err error) {
	iterator := self.db.NewIteratorWithPrefix(blockPrefix)
	for ok := iterator.Seek(blockKey(start)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := utils.DecodeNumber(key[5:13])
		if num >= end {
			break
		}

		var block BlockInfo
		if err = rlp.Decode(bytes.NewReader(iterator.Value()), &block); err != nil {
			log.Error("Exchange Invalid block RLP", "Num", num, "err", err)
			return
		}
		blocks = append(blocks, block)
	}
	return
}

func (self *Exchange) GetRecordsByTxHash(txHash c_type.Uint256) (records []Utxo, err error) {
	data, err := self.db.Get(txKey(txHash))
	if err != nil {
		return
	}
	if err = rlp.Decode(bytes.NewReader(data), &records); err != nil {
		log.Error("Invalid utxos RLP", "txHash", common.Bytes2Hex(txHash[:]), "err", err)
		return
	}
	return
}

func (self *Exchange) GetRecordsByPk(pk *c_type.Uint512, begin, end uint64) (records []Utxo, err error) {
	err = self.iteratorUtxo(pk, begin, end, func(utxo Utxo) {
		records = append(records, utxo)
	})
	return
}

func (self *Exchange) GetRecordsByPkr(pkr c_type.PKr, begin, end uint64) (records []Utxo, err error) {
	account := self.getAccountByPkr(pkr)
	if account == nil {
		err = errors.New("not found PK by pkr")
		return
	}
	pk := account.pk

	err = self.iteratorUtxo(pk, begin, end, func(utxo Utxo) {
		if pkr != utxo.Pkr {
			return
		}
		records = append(records, utxo)
	})
	return
}

func (self *Exchange) GenTxWithSign(param prepare.PreTxParam) (pretx *txtool.GTxParam, tx *txtool.GTx, e error) {
	if self == nil {
		e = errors.New("exchange instance is nil")
		return
	}
	var roots prepare.Utxos
	if roots, e = prepare.SelectUtxos(&param, self); e != nil {
		return
	}

	var account *Account
	if value, ok := self.accounts.Load(param.From); ok {
		account = value.(*Account)
	} else {
		e = errors.New("not found Pk")
		return
	}

	if param.RefundTo == nil {
		if param.RefundTo = self.DefaultRefundTo(&param.From); param.RefundTo == nil {
			e = errors.New("can not find default refund to")
			return
		}
	}

	bparam := prepare.BeforeTxParam{
		param.Fee,
		*param.GasPrice,
		roots,
		*param.RefundTo,
		param.Receptions,
		param.Cmds,
	}

	if pretx, tx, e = self.genTx(account, &bparam); e != nil {
		log.Error("Exchange genTx", "error", e)
		return
	}
	tx.Hash = tx.Tx.ToHash()
	log.Info("Exchange genTx success")
	return
}

func (self *Exchange) getAccountByPk(pk c_type.Uint512) *Account {
	if value, ok := self.accounts.Load(pk); ok {
		return value.(*Account)
	}
	return nil
}

func (self *Exchange) getAccountByPkr(pkr c_type.PKr) (a *Account) {
	self.accounts.Range(func(pk, value interface{}) bool {
		account := value.(*Account)
		if superzk.IsMyPKr(account.tk, &pkr) {
			a = account
			return false
		}
		return true
	})
	return
}

func (self *Exchange) ClearTxParam(txParam *txtool.GTxParam) (count int) {
	if self == nil {
		return
	}
	if txParam == nil {
		return
	}
	for _, in := range txParam.Ins {
		count += self.ClearUsedFlagForRoot(in.Out.Root)
	}
	return
}

func (self *Exchange) genTx(account *Account, param *prepare.BeforeTxParam) (txParam *txtool.GTxParam, tx *txtool.GTx, e error) {
	if txParam, e = self.buildTxParam(param); e != nil {
		return
	}

	var seed *address.Seed
	if seed, e = account.wallet.GetSeed(); e != nil {
		self.ClearTxParam(txParam)
		return
	}

	sk := superzk.Seed2Sk(seed.SeedToUint256())
	gtx, err := flight.SignTx(&sk, txParam)
	if err != nil {
		self.ClearTxParam(txParam)
		e = err
		return
	} else {
		tx = &gtx
		return
	}
}

func (self *Exchange) commitTx(tx *txtool.GTx) (err error) {
	gasPrice := big.Int(tx.GasPrice)
	gas := uint64(tx.Gas)
	signedTx := types.NewTxWithGTx(gas, &gasPrice, &tx.Tx)
	log.Info("Exchange commitTx", "txhash", signedTx.Hash().String())
	err = self.txPool.AddLocal(signedTx)
	return err
}

func (self *Exchange) iteratorUtxo(Pk *c_type.Uint512, begin, end uint64, handler HandleUtxoFunc) (e error) {
	var pk c_type.Uint512
	if Pk != nil {
		pk = *Pk
	}
	iterator := self.db.NewIteratorWithPrefix(utxoPrefix)
	for ok := iterator.Seek(utxoKey(begin, pk)); ok; ok = iterator.Next() {
		key := iterator.Key()
		num := utils.DecodeNumber(key[4:12])
		if num >= end {
			break
		}
		copy(pk[:], key[12:76])

		if Pk != nil && *Pk != pk {
			continue
		}

		value := iterator.Value()
		roots := []c_type.Uint256{}
		if err := rlp.Decode(bytes.NewReader(value), &roots); err != nil {
			log.Error("Invalid roots RLP", "accoutKey", common.Bytes2Hex(pk[:]), "blockNumber", num, "err", err)
			e = err
			return
		}
		for _, root := range roots {
			if utxo, err := self.getUtxo(root); err != nil {
				return
			} else {
				handler(utxo)
			}
		}
	}

	return
}

var ignorePKr = common.Base58ToAddress("i3zesDa26i7jAtkR2fBYBeZsoQ7NAJxQNCsbgwvaWap3HVDGmvzsQSLqTZRyadswzBoC4edWYJzejyY6AXVhGkcqFYVvVPH1w5vHfvbazp1ReQ5Wa9qi15UPAwztrxe9oJQ").ToPKr()

func (self *Exchange) getUtxo(root c_type.Uint256) (utxo Utxo, e error) {
	data, err := self.db.Get(rootKey(root))
	if err != nil {
		return
	}
	if err := rlp.Decode(bytes.NewReader(data), &utxo); err != nil {
		log.Error("Exchange Invalid utxo RLP", "root", common.Bytes2Hex(root[:]), "err", err)
		e = err
		return
	}

	if utxo.Pkr == *ignorePKr {
		utxo.Ignore = true
	}

	if value, ok := self.usedFlag.Load(utxo.Root); ok {
		utxo.flag = value.(int)
	}
	return
}

func (self *Exchange) findUtxosByTicket(pk *c_type.Uint512, tickets []assets.Ticket) (utxos []Utxo, remain map[c_type.Uint256]c_type.Uint256) {
	remain = map[c_type.Uint256]c_type.Uint256{}
	for _, ticket := range tickets {
		remain[ticket.Value] = ticket.Category
		prefix := append(pkPrefix, append(pk[:], ticket.Value[:]...)...)
		iterator := self.db.NewIteratorWithPrefix(prefix)
		if iterator.Next() {
			key := iterator.Key()
			var root c_type.Uint256
			copy(root[:], key[98:130])

			if utxo, err := self.getUtxo(root); err == nil {
				if utxo.Ignore {
					continue
				}
				if utxo.Asset.Tkt != nil && utxo.Asset.Tkt.Category == ticket.Category {
					if _, ok := self.usedFlag.Load(utxo.Root); !ok {
						utxos = append(utxos, utxo)
						delete(remain, ticket.Value)
					}
				}
			}
		}
	}
	return
}

func (self *Exchange) findUtxos(pk *c_type.Uint512, currency string, amount *big.Int) (utxos []Utxo, remain *big.Int) {
	remain = new(big.Int).Set(amount)

	currency = strings.ToUpper(currency)
	prefix := append(pkPrefix, append(pk[:], common.LeftPadBytes([]byte(currency), 32)...)...)
	iterator := self.db.NewIteratorWithPrefix(prefix)

	for iterator.Next() {
		key := iterator.Key()
		var root c_type.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if utxo.Ignore {
				continue
			}
			if utxo.Asset.Tkn != nil {
				if _, ok := self.usedFlag.Load(utxo.Root); !ok {
					utxos = append(utxos, utxo)
					remain.Sub(remain, utxo.Asset.Tkn.Value.ToIntRef())
					if remain.Sign() <= 0 {
						break
					}
				}
			}
		}
	}
	return
}

func DecOuts(outs []txtool.Out, skr *c_type.PKr) (douts []txtool.DOut) {
	tk := c_type.Tk{}
	copy(tk[:], skr[:])
	tdouts := flight.DecOut(&tk, outs)
	for _, tdout := range tdouts {
		ot := txtool.DOut{
			Asset: tdout.Asset,
			Memo:  tdout.Memo,
		}
		if len(tdout.Nils) > 0 {
			ot.Nil = tdout.Nils[0]
		}
		douts = append(douts, ot)
	}
	return
}

type uint64Slice []uint64

func (c uint64Slice) Len() int {
	return len(c)
}
func (c uint64Slice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c uint64Slice) Less(i, j int) bool {
	return c[i] < c[j]
}

var fetchCount = uint64(5000)

func (self *Exchange) fetchBlockInfo() {
	if txtool.Ref_inst.Bc == nil || !txtool.Ref_inst.Bc.IsValid() {
		return
	}
	for {
		indexs := map[uint64][]c_type.Uint512{}
		orders := uint64Slice{}
		self.numbers.Range(func(key, value interface{}) bool {
			pk := key.(c_type.Uint512)
			num := value.(uint64)
			if list, ok := indexs[num]; ok {
				indexs[num] = append(list, pk)
			} else {
				indexs[num] = []c_type.Uint512{pk}
				orders = append(orders, num)
			}
			return true
		})
		if orders.Len() == 0 {
			return
		}

		sort.Sort(orders)
		start := orders[0]
		end := start + fetchCount
		if orders.Len() > 1 {
			end = orders[1]
		}

		accountKeys := indexs[start]
		for end > start {
			count := fetchCount
			if end-start < fetchCount {
				count = end - start
			}
			if count == 0 {
				return
			}
			if self.fetchAndIndexUtxo(start, count, accountKeys) < int(count) {
				return
			}
			start += count
		}
	}
}

func (self *Exchange) fetchAndIndexUtxo(start, countBlock uint64, pks []c_type.Uint512) (count int) {

	blocks, err := flight.SRI_Inst.GetBlocksInfo(start, countBlock)
	if err != nil {
		log.Info("Exchange GetBlocksInfo", "error", err)
		return
	}

	if len(blocks) == 0 {
		return
	}

	utxosMap := map[PkKey][]Utxo{}
	nilsMap := map[c_type.Uint256]Utxo{}
	nils := []c_type.Uint256{}
	blockMap := map[uint64]*BlockInfo{}
	for _, block := range blocks {
		num := uint64(block.Num)
		utxos := []Utxo{}
		for _, out := range block.Outs {
			pkr := out.State.OS.ToPKr()

			if pkr == nil {
				continue
			}

			account, ok := self.ownPkr(pks, *pkr)
			// log.Info("index", ">>>>", account.wallet.Accounts()[0].Address.String())
			if !ok {
				continue
			}

			key := PkKey{key: *account.pk, Num: out.State.Num}
			dout := DecOuts([]txtool.Out{out}, &account.skr)[0]

			if dout.Nil == c_type.Empty_Uint256 {
				continue
			}

			utxo := Utxo{Pkr: *pkr, Root: out.Root, Nil: dout.Nil, TxHash: out.State.TxHash, Num: out.State.Num, Asset: dout.Asset, IsZ: out.State.OS.IsZero()}
			nilsMap[utxo.Root] = utxo
			nilsMap[utxo.Nil] = utxo

			if list, ok := utxosMap[key]; ok {
				utxosMap[key] = append(list, utxo)
			} else {
				utxosMap[key] = []Utxo{utxo}
			}
			utxos = append(utxos, utxo)
		}

		if len(utxos) > 0 {
			blockMap[num] = &BlockInfo{Num: num, Hash: block.Hash, Outs: utxos}
		}

		if len(block.Nils) > 0 {
			roots := []c_type.Uint256{}
			for _, Nil := range block.Nils {
				var utxo Utxo
				if value, ok := nilsMap[Nil]; ok {
					utxo = value
				} else {
					value, _ := self.db.Get(nilKey(Nil))
					if value != nil {
						var root c_type.Uint256
						copy(root[:], value[98:130])

						if utxo, err = self.getUtxo(root); err != nil {
							continue
						} else {
							var pk c_type.Uint512
							copy(pk[:], value[2:66])
						}
					} else {
						continue
					}
				}
				nils = append(nils, Nil)
				roots = append(roots, utxo.Root)
			}
			if len(roots) > 0 {
				if blockInfo, ok := blockMap[num]; ok {
					blockInfo.Ins = roots
				} else {
					blockMap[num] = &BlockInfo{Num: num, Hash: block.Hash, Ins: roots}
				}
			}
		}
	}

	batch := self.db.NewBatch()

	self.indexPkgs(pks, batch, blocks)

	var roots []c_type.Uint256
	if len(utxosMap) > 0 || len(nils) > 0 {
		if roots, err = self.indexBlocks(batch, utxosMap, blockMap, nils); err != nil {
			log.Error("indexBlocks ", "error", err)
			return
		}
	}

	count = len(blocks)
	num := uint64(blocks[count-1].Num) + 1
	// "NUM"+PK  => Num
	data := utils.EncodeNumber(num)
	for _, pk := range pks {
		batch.Put(numKey(pk), data)
	}

	err = batch.Write()
	if err == nil {
		for _, pk := range pks {
			self.numbers.Store(pk, num)
		}
	}

	for _, root := range roots {
		self.usedFlag.Delete(root)
	}
	log.Info("Exchange indexed", "blockNumber", num-1)
	return
}

func (self *Exchange) indexBlocks(batch decedb.Batch, utxosMap map[PkKey][]Utxo, blockMap map[uint64]*BlockInfo, nils []c_type.Uint256) (delRoots []c_type.Uint256, err error) {
	ops := map[string]string{}

	for num, blockInfo := range blockMap {
		data, e := rlp.EncodeToBytes(&blockInfo)
		if e != nil {
			err = e
			return
		}
		batch.Put(blockKey(num), data)
	}

	txMap := map[c_type.Uint256][]Utxo{}
	for key, list := range utxosMap {
		roots := []c_type.Uint256{}
		for _, utxo := range list {
			data, e := rlp.EncodeToBytes(&utxo)
			if e != nil {
				err = e
				return
			}

			// "ROOT" + root
			batch.Put(rootKey(utxo.Root), data)
			// nil => root
			batch.Put(nilToRootKey(utxo.Nil), utxo.Root[:])

			var pkKeys []byte
			if utxo.Asset.Tkn != nil {
				// "PK" + PK + currency + root
				pkKey := utxoPkKey(key.key, utxo.Asset.Tkn.Currency[:], &utxo.Root)
				ops[common.Bytes2Hex(pkKey)] = common.Bytes2Hex([]byte{0})
				pkKeys = append(pkKeys, pkKey...)
			}

			if utxo.Asset.Tkt != nil {
				// "PK" + PK + tkt + root
				pkKey := utxoPkKey(key.key, utxo.Asset.Tkt.Value[:], &utxo.Root)
				ops[common.Bytes2Hex(pkKey)] = common.Bytes2Hex([]byte{0})
				pkKeys = append(pkKeys, pkKey...)
			}
			// "PK" + PK + currency + root => 0

			// "NIL" + PK + tkt + root => "PK" + PK + currency + root
			nilkey := nilKey(utxo.Nil)
			rootkey := nilKey(utxo.Root)

			// "NIL" +nil/root => pkKey
			ops[common.Bytes2Hex(nilkey)] = common.Bytes2Hex(pkKeys)
			ops[common.Bytes2Hex(rootkey)] = common.Bytes2Hex(pkKeys)

			roots = append(roots, utxo.Root)

			if list, ok := txMap[utxo.TxHash]; ok {
				txMap[utxo.TxHash] = append(list, utxo)
			} else {
				txMap[utxo.TxHash] = []Utxo{utxo}
			}

			// log.Info("Index add", "PK", base58.EncodeToString(key.PK[:]), "Nil", common.Bytes2Hex(utxo.Nil[:]), "root", common.Bytes2Hex(utxo.Root[:]), "Value", utxo.Asset.Tkn.Value)
		}

		data, e := rlp.EncodeToBytes(&roots)
		if e != nil {
			err = e
			return
		}
		// blockNumber + PK => [roots]
		batch.Put(utxoKey(key.Num, key.key), data)

		if account := self.getAccountByPk(key.key); account != nil {
			account.isChanged = true
		}
	}

	for txHash, list := range txMap {
		key := txKey(txHash)
		data, e := self.db.Get(key)
		var records []Utxo
		if e == nil {
			if e = rlp.Decode(bytes.NewReader(data), &records); e != nil {
				err = e
				log.Error("Invalid utxos RLP", "txHash", common.Bytes2Hex(txHash[:]), "err", e)
				return
			}
		}
		records = append(records, list...)
		data, err = rlp.EncodeToBytes(&records)
		if err != nil {
			return nil, err
		}
		batch.Put(key, data)
	}

	for _, Nil := range nils {

		var pk c_type.Uint512
		key := nilKey(Nil)
		hex := common.Bytes2Hex(key)
		if value, ok := ops[hex]; ok {
			delete(ops, hex)
			if len(value) == 260 {
				delete(ops, value)
			} else {
				delete(ops, value[0:260])
				delete(ops, value[260:])
			}

			var root c_type.Uint256
			copy(root[:], value[98:130])
			delete(ops, common.Bytes2Hex(nilKey(root)))
			// self.usedFlag.Delete(root)
			delRoots = append(delRoots, root)

			copy(pk[:], value[2:66])
		} else {
			value, _ := self.db.Get(key)
			if value != nil {
				if len(value) == 130 {
					batch.Delete(value)
				} else {
					batch.Delete(value[0:130])
					batch.Delete(value[130:260])
				}
				batch.Delete(nilKey(Nil))

				var root c_type.Uint256
				copy(root[:], value[98:130])
				batch.Delete(nilKey(root))
				// self.usedFlag.Delete(root)
				delRoots = append(delRoots, root)

				copy(pk[:], value[2:66])
			}
		}

		if account := self.getAccountByPk(pk); account != nil {
			account.isChanged = true
		}
	}

	for key, value := range ops {
		batch.Put(common.Hex2Bytes(key), common.Hex2Bytes(value))
	}

	return
}

func (self *Exchange) ownPkr(pks []c_type.Uint512, pkr c_type.PKr) (account *Account, ok bool) {
	for _, pk := range pks {
		value, ok := self.accounts.Load(pk)
		if !ok {
			continue
		}
		account = value.(*Account)
		if account.balancePkr != nil {
			if pkr == *account.balancePkr {
				return account, true
			}
		} else {
			if superzk.IsMyPKr(account.tk, &pkr) {
				return account, true
			}
		}

	}
	return
}

type MergeUtxos struct {
	list   UtxoList
	zcount int
	ocount int
	//amount  big.Int
	//tickets map[c_type.Uint256]c_type.Uint256
}

var default_fee_value = new(big.Int).Mul(big.NewInt(25000), big.NewInt(1000000000))

func (self *Exchange) getMergeUtxos(from *c_type.Uint512, currency string, zcount int, left int, icount int) (mu MergeUtxos, e error) {
	if zcount > 400 {
		e = errors.New("zout count must <= 400")
	}
	if icount <= 0 {
		icount = 1000
	}
	ck := assets.NewCKState(true, &assets.Token{utils.CurrencyToUint256("DECE"), utils.U256(*default_fee_value)})
	prefix := utxoPkKey(*from, common.LeftPadBytes([]byte(currency), 32), nil)
	iterator := self.db.NewIteratorWithPrefix(prefix)
	outxos := UtxoList{}
	zutxos := UtxoList{}
	for iterator.Next() {
		key := iterator.Key()
		var root c_type.Uint256
		copy(root[:], key[98:130])

		if utxo, err := self.getUtxo(root); err == nil {
			if utxo.Ignore {
				continue
			}
			if _, ok := self.usedFlag.Load(utxo.Root); !ok {
				if utxo.IsZ {
					zutxos = append(zutxos, utxo)
				} else {
					outxos = append(outxos, utxo)
				}

			}
		}
		if zutxos.Len() >= zcount+left {
			break
		}
		if outxos.Len()+zutxos.Len() >= icount+left {
			break
		}
	}
	if outxos.Len() >= icount {
		zutxos = UtxoList{}
	}
	mu.ocount = outxos.Len()
	mu.zcount = zutxos.Len()
	utxos := append(zutxos, outxos...)
	if utxos.Len() <= left {
		e = fmt.Errorf("no need to merge the account, utxo count == %v", utxos.Len())
		return
	}
	sort.Sort(utxos)
	mu.list = utxos[0 : utxos.Len()-(left-1)]
	for _, utxo := range mu.list {
		ck.AddIn(&utxo.Asset)
	}

	for _, tkn := range ck.Tkns() {
		if utxos, r := self.findUtxos(from, utils.BytesToCurrency(tkn.Currency[:]), tkn.Value.ToInt()); r == nil || r.Sign() > 0 {
			e = errors.New("No enough DECE coins for fee")
			return
		} else {
			mu.list = append(mu.list, utxos...)
		}
	}
	return
}

type MergeParam struct {
	From     c_type.Uint512
	To       *c_type.PKr
	Currency string
	Zcount   uint64
	Left     uint64
	Icount   uint64
}

func (self *Exchange) GenMergeTx(mp *MergeParam) (txParam *txtool.GTxParam, e error) {
	account := self.getAccountByPk(mp.From)
	if account == nil {
		e = errors.New("account is nil")
		return
	}
	if mp.To == nil {
		mp.To = account.wallet.Accounts()[0].GetDefaultPkr(1).NewRef()
	}
	var mu MergeUtxos
	if mu, e = self.getMergeUtxos(account.pk, mp.Currency, int(mp.Zcount), int(mp.Left), int(mp.Icount)); e != nil {
		return
	}

	bytes := common.LeftPadBytes([]byte(mp.Currency), 32)
	var Currency c_type.Uint256
	copy(Currency[:], bytes[:])

	ck := assets.NewCKState(false, &assets.Token{utils.CurrencyToUint256("DECE"), utils.U256(*default_fee_value)})

	for _, utxo := range mu.list {
		ck.AddIn(&utxo.Asset)
	}

	receptions := []prepare.Reception{}

	for _, utxo := range ck.Tkns() {
		receptions = append(receptions, prepare.Reception{
			Addr: *mp.To,
			Asset: assets.Asset{
				Tkn: &assets.Token{
					Currency: utxo.Currency,
					Value:    utxo.Value,
				},
			},
		})
	}

	for _, utxo := range ck.Tkts() {
		receptions = append(receptions, prepare.Reception{
			Addr: *mp.To,
			Asset: assets.Asset{
				Tkt: &assets.Ticket{
					Category: utxo.Category,
					Value:    utxo.Value,
				},
			},
		})
	}

	bparam := prepare.BeforeTxParam{
		assets.Token{
			utils.CurrencyToUint256("DECE"),
			utils.U256(*default_fee_value),
		},
		*big.NewInt(1000000000),
		mu.list.Roots(),
		*mp.To,
		receptions,
		prepare.Cmds{},
	}

	txParam, e = self.buildTxParam(&bparam)
	if e != nil {
		return
	}
	return
}

func (self *Exchange) Merge(pk *c_type.Uint512, currency string, force bool) (count int, txhash c_type.Uint256, e error) {
	account := self.getAccountByPk(*pk)
	if account == nil {
		e = errors.New("account is nil")
		return
	}

	seed, err := account.wallet.GetSeed()
	if err != nil || seed == nil {
		e = errors.New("account is locked")
		return
	}

	var mu MergeUtxos
	if mu, e = self.getMergeUtxos(account.pk, currency, 100, 50, 0); e != nil {
		return
	}

	if mu.zcount >= 100 || mu.ocount >= 2400 || time.Now().After(account.nextMergeTime) || force {

		count = mu.list.Len()
		bytes := common.LeftPadBytes([]byte(currency), 32)
		var Currency c_type.Uint256
		copy(Currency[:], bytes[:])

		ck := assets.NewCKState(false, &assets.Token{utils.CurrencyToUint256("DECE"), utils.U256(*default_fee_value)})

		for _, utxo := range mu.list {
			ck.AddIn(&utxo.Asset)
		}

		receptions := []prepare.Reception{}

		for _, utxo := range ck.Tkns() {
			receptions = append(receptions, prepare.Reception{
				Addr: account.mainPkr,
				Asset: assets.Asset{
					Tkn: &assets.Token{
						Currency: utxo.Currency,
						Value:    utxo.Value,
					},
				},
			})
		}

		for _, utxo := range ck.Tkts() {
			receptions = append(receptions, prepare.Reception{
				Addr: account.mainPkr,
				Asset: assets.Asset{
					Tkt: &assets.Ticket{
						Category: utxo.Category,
						Value:    utxo.Value,
					},
				},
			})
		}

		bparam := prepare.BeforeTxParam{
			assets.Token{
				utils.CurrencyToUint256("DECE"),
				utils.U256(*default_fee_value),
			},
			*big.NewInt(1000000000),
			mu.list.Roots(),
			account.mainPkr,
			receptions,
			prepare.Cmds{},
		}

		pretx, gtx, err := self.genTx(account, &bparam)
		if err != nil {
			account.nextMergeTime = time.Now().Add(time.Hour * 6)
			e = err
			return
		}
		txhash = gtx.Hash
		if err := self.commitTx(gtx); err != nil {
			account.nextMergeTime = time.Now().Add(time.Hour * 6)
			self.ClearTxParam(pretx)
			e = err
			return
		}
		if mu.list.Len() < 100 {
			account.nextMergeTime = time.Now().Add(time.Hour * 6)
		}
		return
	} else {
		e = fmt.Errorf("no need to merge the account, utxo count == %v", mu.list.Len())
		return
	}
}

func (self *Exchange) merge() {
	if txtool.Ref_inst.Bc == nil || !txtool.Ref_inst.Bc.IsValid() {
		return
	}
	self.accounts.Range(func(key, value interface{}) bool {
		account := value.(*Account)
		if count, txhash, err := self.Merge(account.pk, "DECE", false); err != nil {
			log.Error("autoMerge fail", "accountKey", *utils.Base58Encode(account.pk[:]), "count", count, "error", err)
		} else {
			log.Info("autoMerge succ", "accountKey", *utils.Base58Encode(account.pk[:]), "tx", hexutil.Encode(txhash[:]), "count", count)
		}
		return true
	})

}

var (
	numPrefix       = []byte("NUM")
	balancPkrPrefix = []byte("BALANCPKR")
	pkPrefix        = []byte("PK")
	utxoPrefix      = []byte("UTXO")
	rootPrefix      = []byte("ROOT")
	nilPrefix       = []byte("NIL")

	blockPrefix   = []byte("BLOCK")
	outUtxoPrefix = []byte("OUTUTXO")
	txPrefix      = []byte("TX")
	nilRootPrefix = []byte("NOILTOROOT")
)

func nilToRootKey(nil c_type.Uint256) []byte {
	return append(nilRootPrefix, nil[:]...)
}

func txKey(txHash c_type.Uint256) []byte {
	return append(txPrefix, txHash[:]...)
}

func blockKey(number uint64) []byte {
	return append(blockPrefix, utils.EncodeNumber(number)...)
}

func numKey(pk c_type.Uint512) []byte {
	return append(numPrefix, pk[:]...)
}

func balancPkrKey(pk c_type.Uint512) []byte {
	return append(balancPkrPrefix, pk[:]...)
}

func nilKey(nil c_type.Uint256) []byte {
	return append(nilPrefix, nil[:]...)
}

func rootKey(root c_type.Uint256) []byte {
	return append(rootPrefix, root[:]...)
}

// func outUtxoKey(number uint64, pk c_type.Uint512) []byte {
//	return append(outUtxoPrefix, append(encodeNumber(number), pk[:]...)...)
// }

// utxoKey = PK + currency +root
func utxoPkKey(pk c_type.Uint512, currency []byte, root *c_type.Uint256) []byte {
	key := append(pkPrefix, pk[:]...)
	if len(currency) > 0 {
		key = append(key, currency...)
	}
	if root != nil {
		key = append(key, root[:]...)
	}
	return key
}

func utxoKey(number uint64, pk c_type.Uint512) []byte {
	return append(utxoPrefix, append(utils.EncodeNumber(number), pk[:]...)...)
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
