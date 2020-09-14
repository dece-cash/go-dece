// Copyright 2017 The go-ethereum Authors
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

package keystore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dece-cash/go-dece/common/address"
	"github.com/dece-cash/go-dece/czero/superzk"

	"github.com/dece-cash/go-dece/accounts"
	"github.com/dece-cash/go-dece/log"
	mapset "github.com/deckarep/golang-set"
)

// Minimum amount of time between cache reloads. This limit applies if the platform does
// not support change notifications. It also applies if the keystore directory does not
// exist yet, the code will attempt to create a watcher at most this often.
const minReloadInterval = 2 * time.Second

type accountsByTag []accountByTag

type accountByTag struct {
	accountByURL accounts.Account
	update       bool
}

func (s accountsByTag) Len() int { return len(s) }
func (s accountsByTag) Less(i, j int) bool {
	return s[i].accountByURL.URL.Cmp(s[j].accountByURL.URL) < 0
}
func (s accountsByTag) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// AmbiguousAddrError is returned when attempting to unlock
// an address for which more than one file exists.
type AmbiguousAddrError struct {
	Address address.PKAddress
	Matches []accounts.Account
}

func (err *AmbiguousAddrError) Error() string {
	files := ""
	for i, a := range err.Matches {
		files += a.URL.Path
		if i < len(err.Matches)-1 {
			files += ", "
		}
	}
	return fmt.Sprintf("multiple keys match address (%s)", files)
}

// accountCache is a live index of all accounts in the keystore.
type accountCache struct {
	keydir   string
	watcher  *watcher
	mu       sync.Mutex
	all      accountsByTag
	byAddr   map[address.PKAddress][]accounts.Account
	throttle *time.Timer
	notify   chan struct{}
	fileC    fileCache
}

func newAccountCache(keydir string) (*accountCache, chan struct{}) {
	ac := &accountCache{
		keydir: keydir,
		byAddr: make(map[address.PKAddress][]accounts.Account),
		notify: make(chan struct{}, 1),
		fileC:  fileCache{all: mapset.NewThreadUnsafeSet()},
	}
	ac.watcher = newWatcher(ac)
	return ac, ac.notify
}

func (ac *accountCache) accounts() []accounts.Account {
	ac.maybeReload()
	ac.mu.Lock()
	defer ac.mu.Unlock()
	cpy := []accounts.Account{}
	for _, accT := range ac.all {
		cpy = append(cpy, accT.accountByURL)
	}
	return cpy
}

func (ac *accountCache) accountsByTag() []accountByTag {
	ac.maybeReload()
	ac.mu.Lock()
	defer ac.mu.Unlock()
	cpy := make([]accountByTag, len(ac.all))
	copy(cpy, ac.all)
	return cpy
}

func (ac *accountCache) hasAddress(addr address.PKAddress) bool {
	ac.maybeReload()
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return len(ac.byAddr[addr]) > 0
}

func (ac *accountCache) add(newAccount accounts.Account, update bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	i := sort.Search(len(ac.all), func(i int) bool { return ac.all[i].accountByURL.URL.Cmp(newAccount.URL) >= 0 })
	if i < len(ac.all) && ac.all[i].accountByURL == newAccount {
		return
	}
	// newAccount is not in the cache.
	ac.all = append(ac.all, accountByTag{accountByURL: accounts.Account{}})
	copy(ac.all[i+1:], ac.all[i:])
	ac.all[i] = accountByTag{newAccount, update}
	ac.byAddr[newAccount.Address] = append(ac.byAddr[newAccount.Address], newAccount)
}

// note: removed needs to be unique here (i.e. both File and Data must be set).
func (ac *accountCache) delete(removed accounts.Account) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.all = removeAccountByTag(ac.all, removed)
	if ba := removeAccount(ac.byAddr[removed.Address], removed); len(ba) == 0 {
		delete(ac.byAddr, removed.Address)
	} else {
		ac.byAddr[removed.Address] = ba
	}
}

// deleteByFile removes an account referenced by the given path.
func (ac *accountCache) deleteByFile(path string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	i := sort.Search(len(ac.all), func(i int) bool { return ac.all[i].accountByURL.URL.Path >= path })

	if i < len(ac.all) && ac.all[i].accountByURL.URL.Path == path {
		removed := ac.all[i].accountByURL
		ac.all = append(ac.all[:i], ac.all[i+1:]...)
		if ba := removeAccount(ac.byAddr[removed.Address], removed); len(ba) == 0 {
			delete(ac.byAddr, removed.Address)
		} else {
			ac.byAddr[removed.Address] = ba
		}
	}
}

func removeAccount(slice []accounts.Account, elem accounts.Account) []accounts.Account {
	for i := range slice {
		if slice[i] == elem {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func removeAccountByTag(slice []accountByTag, elem accounts.Account) []accountByTag {
	for i := range slice {
		if slice[i].accountByURL == elem {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// find returns the cached account for address if there is a unique match.
// The exact matching rules are explained by the documentation of accounts.AccountAddress.
// Callers must hold ac.mu.
func (ac *accountCache) find(a accounts.Account) (accounts.Account, error) {
	// Limit search to address candidates if possible.
	matches := []accounts.Account{}

	for _, accT := range ac.all {
		matches = append(matches, accT.accountByURL)
	}

	if (a.Address != address.PKAddress{}) {
		matches = ac.byAddr[a.Address]
	}
	if a.URL.Path != "" {
		// If only the basename is specified, complete the path.
		if !strings.ContainsRune(a.URL.Path, filepath.Separator) {
			a.URL.Path = filepath.Join(ac.keydir, a.URL.Path)
		}
		for i := range matches {
			if matches[i].URL == a.URL {
				return matches[i], nil
			}
		}
		if (a.Address == address.PKAddress{}) {
			return accounts.Account{}, ErrNoMatch
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return accounts.Account{}, ErrNoMatch
	default:
		err := &AmbiguousAddrError{Address: a.Address, Matches: make([]accounts.Account, len(matches))}
		copy(err.Matches, matches)
		//sort.Sort(accountsByURL(err.Matches))
		return accounts.Account{}, err
	}
}

func (ac *accountCache) maybeReload() {
	ac.mu.Lock()

	if ac.watcher.running {
		ac.mu.Unlock()
		return // A watcher is running and will keep the cache up-to-date.
	}
	if ac.throttle == nil {
		ac.throttle = time.NewTimer(0)
	} else {
		select {
		case <-ac.throttle.C:
		default:
			ac.mu.Unlock()
			return // The cache was reloaded recently.
		}
	}
	// No watcher running, start it.
	ac.watcher.start()
	ac.throttle.Reset(minReloadInterval)
	ac.mu.Unlock()
	ac.scanAccounts()
}

func (ac *accountCache) close() {
	ac.mu.Lock()
	ac.watcher.close()
	if ac.throttle != nil {
		ac.throttle.Stop()
	}
	if ac.notify != nil {
		close(ac.notify)
		ac.notify = nil
	}
	ac.mu.Unlock()
}

// scanAccounts checks if any changes have occurred on the filesystem, and
// updates the account cache accordingly
func (ac *accountCache) scanAccounts() error {
	// Scan the entire folder metadata for file changes
	creates, deletes, updates, err := ac.fileC.scan(ac.keydir)
	if err != nil {
		log.Debug("Failed to reload keystore contents", "err", err)
		return err
	}
	if creates.Cardinality() == 0 && deletes.Cardinality() == 0 && updates.Cardinality() == 0 {
		return nil
	}
	// Create a helper method to scan the contents of the key files
	var (
		buf = new(bufio.Reader)
		key struct {
			Address string `json:"address"`
			Tk      string `json:"tk"`
			At      uint64 `json:"at"`
		}
	)
	readAccount := func(path string) *accounts.Account {
		fd, err := os.Open(path)
		if err != nil {
			log.Trace("Failed to open keystore file", "path", path, "err", err)
			return nil
		}
		defer fd.Close()
		buf.Reset(fd)
		// Parse the address.
		key.Address = ""
		key.Tk = ""
		key.At = 0
		err = json.NewDecoder(buf).Decode(&key)
		tk := address.Base58ToTk(key.Tk)
		pk, _ := superzk.Tk2Pk(tk.ToTk().NewRef())
		var addr address.PKAddress
		copy(addr[:], pk[:])
		at := key.At

		switch {
		case err != nil:
			log.Debug("Failed to decode keystore key", "path", path, "err", err)
		case (addr == address.PKAddress{}):
			log.Debug("Failed to decode keystore key", "path", path, "err", "missing or zero address")
		default:
			return &accounts.Account{Address: addr, Tk: tk, URL: accounts.URL{Scheme: KeyStoreScheme, Path: path}, At: at}
		}
		return nil
	}
	// Process all the file diffs
	start := time.Now()

	for _, p := range creates.ToSlice() {
		if a := readAccount(p.(string)); a != nil {
			ac.add(*a, true)
		}
	}

	for _, p := range deletes.ToSlice() {
		ac.deleteByFile(p.(string))
	}
	for _, p := range updates.ToSlice() {
		path := p.(string)
		ac.deleteByFile(path)
		if a := readAccount(path); a != nil {
			ac.add(*a, true)
		}
	}
	end := time.Now()

	select {
	case ac.notify <- struct{}{}:
	default:
	}
	log.Trace("Handled keystore changes", "time", end.Sub(start))
	return nil
}
