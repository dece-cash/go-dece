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

// Package accounts implements high level Sero account management.
package accounts

import (
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/czero/superzk"
	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/common/address"
	"github.com/dece-cash/go-dece/event"
	"github.com/dece-cash/go-dece/zero/utils"
)

// AccountAddress represents an Sero account located at a specific location defined
// by the optional URL field.
type Account struct {
	Address address.PKAddress `json:"address"` // Sero account address derived from the key
	Tk      address.TKAddress `json:"tk"`      // Sero account tk derived from the key
	URL     URL               `json:"url"`     // Optional resource locator within a backend
	At      uint64            `json:"at"`      //account create at blocknum
	Version int               `json:"version"`
}

func (self *Account) GetPkr(rand *c_type.Uint256) c_type.PKr {
	pk := self.Address.ToUint512()
	return superzk.Pk2PKr(&pk, rand)
}
func (self *Account) GetPk() c_type.Uint512 {
	var pk c_type.Uint512
	pk, _ = superzk.Tk2Pk(self.Tk.ToTk().NewRef())
	return pk
}

func (self *Account) GetDefaultPkr(index uint64) c_type.PKr {
	pk := self.Address.ToUint512()
	r := c_type.Uint256{}
	copy(r[:], common.LeftPadBytes(utils.EncodeNumber(index), 32))
	if index == 0 {
		return superzk.Pk2PKr(&pk, nil)
	} else {
		return superzk.Pk2PKr(&pk, &r)
	}
}

func (self *Account) IsMyPk(pk c_type.Uint512) bool {
	pkr := superzk.Pk2PKr(&pk, nil)
	return self.IsMyPkr(pkr)

}

func (self *Account) IsMyPkr(pkr c_type.PKr) bool {
	tk := c_type.Tk{}
	copy(tk[:], self.Tk[:])
	return superzk.IsMyPKr(&tk, &pkr)
}

// Wallet represents a software or hardware wallet that might contain one or more
// accounts (derived from the same seed).
type Wallet interface {
	// URL retrieves the canonical path under which this wallet is reachable. It is
	// user by upper layers to define a sorting order over all wallets from multiple
	// backends.
	URL() URL

	// Status returns a textual status to aid the user in the current state of the
	// wallet. It also returns an error indicating any failure the wallet might have
	// encountered.
	Status() (string, error)

	// Open initializes access to a wallet instance. It is not meant to unlock or
	// decrypt account keys, rather simply to establish a connection to hardware
	// wallets and/or to access derivation seeds.
	//
	// The passphrase parameter may or may not be used by the implementation of a
	// particular wallet instance. The reason there is no passwordless open method
	// is to strive towards a uniform wallet handling, oblivious to the different
	// backend providers.
	//
	// Please note, if you open a wallet, you must close it to release any allocated
	// resources (especially important when working with hardware wallets).
	Open(passphrase string) error

	// Close releases any resources held by an open wallet instance.
	Close() error

	// Accounts retrieves the list of signing accounts the wallet is currently aware
	// of. For hierarchical deterministic wallets, the list will not be exhaustive,
	// rather only contain the accounts explicitly pinned during account derivation.
	Accounts() []Account

	// Contains returns whether an account is part of this particular wallet or not.
	Contains(account Account) bool

	// Derive attempts to explicitly derive a hierarchical deterministic account at
	// the specified derivation path. If requested, the derived account will be added
	// to the wallet's tracked account list.
	Derive(path DerivationPath, pin bool) (Account, error)

	// SelfDerive sets a base account derivation path from which the wallet attempts
	// to discover non zero accounts and automatically add them to list of tracked
	// accounts.
	//
	// Note, self derivaton will increment the last component of the specified path
	// opposed to decending into a child path to allow discovering accounts starting
	// from non zero components.
	//
	// You can disable automatic account discovery by calling SelfDerive with a nil
	// chain state reader.

	// IsMine return whether an once address is mine or not
	IsMine(pkr c_type.PKr) bool

	AddressUnlocked(account Account) (bool, error)

	GetSeed() (*address.Seed, error)

	GetSeedWithPassphrase(passphrase string) (*address.Seed, error)
}

// Backend is a "wallet provider" that may contain a batch of accounts they can
// sign transactions with and upon request, do so.
type Backend interface {
	// Wallets retrieves the list of wallets the backend is currently aware of.
	//
	// The returned wallets are not opened by default. For software HD wallets this
	// means that no base seeds are decrypted, and for hardware wallets that no actual
	// connection is established.
	//
	// The resulting wallet list will be sorted alphabetically based on its internal
	// URL assigned by the backend. Since wallets (especially hardware) may come and
	// go, the same wallet might appear at a different positions in the list during
	// subsequent retrievals.
	Wallets() []Wallet

	// Subscribe creates an async subscription to receive notifications when the
	// backend detects the arrival or departure of a wallet.
	Subscribe(sink chan<- WalletEvent) event.Subscription
}

// WalletEventType represents the different event types that can be fired by
// the wallet subscription subsystem.
type WalletEventType int

const (
	// WalletArrived is fired when a new wallet is detected either via USB or via
	// a filesystem event in the keystore.
	WalletArrived WalletEventType = iota

	// WalletOpened is fired when a wallet is successfully opened with the purpose
	// of starting any background processes such as automatic key derivation.
	WalletOpened

	// WalletDropped
	WalletDropped
)

// WalletEvent is an event fired by an account backend when a wallet arrival or
// departure is detected.
type WalletEvent struct {
	Wallet Wallet          // Wallet instance arrived or departed
	Kind   WalletEventType // Event type that happened in the system
}
