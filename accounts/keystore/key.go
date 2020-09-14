// Copyright 2014 The go-ethereum Authors
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
	"crypto/ecdsa"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/dece-cash/go-dece/common/address"

	"github.com/dece-cash/go-dece/czero/c_type"

	bip39 "github.com/tyler-smith/go-bip39"

	"github.com/pborman/uuid"
	"github.com/dece-cash/go-dece/accounts"
	"github.com/dece-cash/go-dece/crypto"
)

type Key struct {
	Id uuid.UUID // Version 4 "random" for unique id not derived from key data
	// to simplify lookups we also store the address
	Address address.PKAddress

	Tk address.TKAddress
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	PrivateKey *ecdsa.PrivateKey

	At uint64
}

type keyStore interface {
	// Loads and decrypts the key from disk.
	GetKey(address address.PKAddress, filename string, auth string) (*Key, error)
	// Writes and encrypts the key.
	StoreKey(filename string, k *Key, auth string) error
	// Joins filename with the key directory unless it is already absolute.
	JoinPath(filename string) string
}

type encryptedKeyJSONV1 struct {
	Address string     `json:"address"`
	Tk      string     `json:"tk"`
	Crypto  cryptoJSON `json:"crypto"`
	Id      string     `json:"id"`
	At      uint64     `json:"at"`
}

type cryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

func newKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey, at uint64) *Key {
	id := uuid.NewRandom()
	tk := crypto.PrivkeyToTk(privateKeyECDSA)
	key := &Key{
		Id:         id,
		Address:    tk.ToPk(),
		Tk:         tk,
		PrivateKey: privateKeyECDSA,
		At:         at,
	}
	return key
}

func newKeyFromTk(tk *c_type.Tk, at uint64) *Key {
	id := uuid.NewRandom()
	tkaddress := address.TKAddress{}
	copy(tkaddress[:], tk[:])

	key := &Key{
		Id:      id,
		Address: tkaddress.ToPk(),
		Tk:      tkaddress,
		At:      at,
	}
	return key
}

func newKey(rand io.Reader, at uint64) (*Key, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand)
	if err != nil {
		return nil, err
	}
	return newKeyFromECDSA(privateKeyECDSA, at), nil
}

func storeNewKey(ks keyStore, rand io.Reader, auth string, at uint64) (*Key, accounts.Account, error) {
	key, err := newKey(rand, at)
	if err != nil {
		return nil, accounts.Account{}, err
	}
	a := accounts.Account{Address: key.Address, Tk: key.Tk, URL: accounts.URL{Scheme: KeyStoreScheme, Path: ks.JoinPath(keyFileName(key.Address))}, At: key.At}
	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
		zeroKey(key.PrivateKey)
		return nil, a, err
	}
	return key, a, err
}

func storeNewKeyWithMnemonic(ks keyStore, auth string, at uint64) (string, *Key, accounts.Account, error) {

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", nil, accounts.Account{}, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", nil, accounts.Account{}, err
	}

	//seed := bip39.NewSeed(mnemonic, "")

	privateKeyECDSA, err := crypto.ToECDSA(entropy)
	if err != nil {
		return "", nil, accounts.Account{}, err
	}

	key := newKeyFromECDSA(privateKeyECDSA, at)
	a := accounts.Account{Address: key.Address, Tk: key.Tk, URL: accounts.URL{Scheme: KeyStoreScheme, Path: ks.JoinPath(keyFileName(key.Address))}, At: key.At}
	if err := ks.StoreKey(a.URL.Path, key, auth); err != nil {
		zeroKey(key.PrivateKey)
		return "", nil, a, err
	}

	return mnemonic, key, a, err
}

func writeKeyFile(file string, content []byte) error {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}

// keyFileName implements the naming convention for keyfiles:
// UTC--<created_at UTC ISO8601>-<address hex>
func keyFileName(keyAddr address.PKAddress) string {
	ts := time.Now().UTC()
	return fmt.Sprintf("UTC--%s--%s", toISO8601(ts), keyAddr.String())
}

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}
