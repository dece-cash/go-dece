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

/*

This key store behaves as KeyStorePlain with the difference that
the private key is encrypted and on disk uses another JSON encoding.

The crypto is documented at https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition

*/

package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/dece-cash/go-dece/common/address"

	"github.com/dece-cash/go-dece/accounts"

	"github.com/btcsuite/btcutil/base58"

	"github.com/pborman/uuid"
	"github.com/dece-cash/go-dece/common/math"
	"github.com/dece-cash/go-dece/crypto"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

const (
	keyHeaderKDF = "scrypt"

	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = 1 << 18

	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptP = 1

	// LightScryptN is the N parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptN = 1 << 12

	// LightScryptP is the P parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptP = 6

	scryptR     = 8
	scryptDKLen = 32

	veryLightScryptN = 2
	veryLightScryptP = 1
)

type keyStorePassphrase struct {
	keysDirPath string
	scryptN     int
	scryptP     int
}

func (ks keyStorePassphrase) GetKey(address address.PKAddress, filename, auth string) (*Key, error) {
	// Load the key from the keystore and decrypt its contents

	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	key, err := DecryptKey(keyjson, auth)

	if err != nil {
		return nil, err
	}
	// Make sure we're really operating on the requested key (no swap attacks)
	if key.Address != address {
		return nil, fmt.Errorf("key content mismatch: have account %s, want %s", key.Address.String(), address.String())
	}
	return key, nil
}

// StoreKey generates a key, encrypts with 'auth' and stores in the given directory
func StoreKey(dir, auth string, scryptN, scryptP int, at uint64) (accounts.Account, error) {
	_, a, err := storeNewKey(&keyStorePassphrase{dir, scryptN, scryptP}, rand.Reader, auth, at)
	return a, err
}

func (ks keyStorePassphrase) StoreKey(filename string, key *Key, auth string) error {
	keyjson, err := EncryptKey(key, auth, ks.scryptN, ks.scryptP)
	if err != nil {
		return err
	}
	return writeKeyFile(filename, keyjson)
}

func (ks keyStorePassphrase) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.keysDirPath, filename)
}

// EncryptKey encrypts a key using the specified scrypt parameters into a json
// blob that can be decrypted later on.
func EncryptKey(key *Key, auth string, scryptN, scryptP int) ([]byte, error) {
	var cryptoStruct cryptoJSON
	if key.PrivateKey != nil {
		authArray := []byte(auth)

		salt := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			panic("reading from crypto/rand failed: " + err.Error())
		}
		derivedKey, err := scrypt.Key(authArray, salt, scryptN, scryptR, scryptP, scryptDKLen)
		if err != nil {
			return nil, err
		}
		encryptKey := derivedKey[:16]
		keyBytes := math.PaddedBigBytes(key.PrivateKey.D, 32)

		iv := make([]byte, aes.BlockSize) // 16
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			panic("reading from crypto/rand failed: " + err.Error())
		}
		cipherText, err := aesCTRXOR(encryptKey, keyBytes, iv)
		if err != nil {
			return nil, err
		}
		mac := crypto.Keccak256(derivedKey[16:32], cipherText)

		scryptParamsJSON := make(map[string]interface{}, 5)
		scryptParamsJSON["n"] = scryptN
		scryptParamsJSON["r"] = scryptR
		scryptParamsJSON["p"] = scryptP
		scryptParamsJSON["dklen"] = scryptDKLen
		scryptParamsJSON["salt"] = hex.EncodeToString(salt)

		cipherParamsJSON := cipherparamsJSON{
			IV: hex.EncodeToString(iv),
		}

		cryptoStruct = cryptoJSON{
			Cipher:       "aes-128-ctr",
			CipherText:   hex.EncodeToString(cipherText),
			CipherParams: cipherParamsJSON,
			KDF:          keyHeaderKDF,
			KDFParams:    scryptParamsJSON,
			MAC:          hex.EncodeToString(mac),
		}
	}
	encryptedKeyJSONV1 := encryptedKeyJSONV1{
		Address: base58.Encode(key.Address[:]),
		Tk:      base58.Encode(key.Tk[:]),
		Crypto:  cryptoStruct,
		Id:      key.Id.String(),
		At:      key.At,
	}
	return json.Marshal(encryptedKeyJSONV1)
}

func GetAddress(keyjson []byte) (string, error) {
	// Parse the json into a simple map to fetch the key version
	m := make(map[string]interface{})
	if err := json.Unmarshal(keyjson, &m); err != nil {
		return "", err
	}
	k := new(encryptedKeyJSONV1)
	if err := json.Unmarshal(keyjson, k); err != nil {
		return "", err
	} else {
		return k.Address, nil
	}
}

// DecryptKey decrypts a key from a json blob, returning the private key itself.
func DecryptKey(keyjson []byte, auth string) (*Key, error) {
	// Parse the json into a simple map to fetch the key version
	m := make(map[string]interface{})
	if err := json.Unmarshal(keyjson, &m); err != nil {
		return nil, err
	}
	// Depending on the version try to parse one way or another
	var (
		keyBytes, keyId []byte
		err             error
	)

	k := new(encryptedKeyJSONV1)
	if err := json.Unmarshal(keyjson, k); err != nil {
		return nil, err
	}
	keyBytes, keyId, err = decryptKeyV3(k, auth)

	// Handle any decryption errors and return the key
	if err != nil {
		return nil, err
	}
	key := crypto.ToECDSAUnsafe(keyBytes)
	tk := crypto.PrivkeyToTk(key)
	return &Key{
		Id:         uuid.UUID(keyId),
		Address:    tk.ToPk(),
		Tk:         tk,
		PrivateKey: key,
		At:         k.At,
	}, nil
}

func decryptKeyV3(keyProtected *encryptedKeyJSONV1, auth string) (keyBytes []byte, keyId []byte, err error) {

	if keyProtected.Crypto.CipherText == "" {
		return nil, nil, fmt.Errorf("has no privatekey for tk:%v", keyProtected.Tk)
	}

	if keyProtected.Crypto.Cipher != "aes-128-ctr" {
		return nil, nil, fmt.Errorf("Cipher not supported: %v", keyProtected.Crypto.Cipher)
	}

	keyId = uuid.Parse(keyProtected.Id)
	mac, err := hex.DecodeString(keyProtected.Crypto.MAC)
	if err != nil {
		return nil, nil, err
	}

	iv, err := hex.DecodeString(keyProtected.Crypto.CipherParams.IV)
	if err != nil {
		return nil, nil, err
	}

	cipherText, err := hex.DecodeString(keyProtected.Crypto.CipherText)
	if err != nil {
		return nil, nil, err
	}

	derivedKey, err := getKDFKey(keyProtected.Crypto, auth)
	if err != nil {
		return nil, nil, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, nil, ErrDecrypt
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, nil, err
	}
	return plainText, keyId, err
}

func getKDFKey(cryptoJSON cryptoJSON, auth string) ([]byte, error) {
	authArray := []byte(auth)
	salt, err := hex.DecodeString(cryptoJSON.KDFParams["salt"].(string))
	if err != nil {
		return nil, err
	}
	dkLen := ensureInt(cryptoJSON.KDFParams["dklen"])

	if cryptoJSON.KDF == keyHeaderKDF {
		n := ensureInt(cryptoJSON.KDFParams["n"])
		r := ensureInt(cryptoJSON.KDFParams["r"])
		p := ensureInt(cryptoJSON.KDFParams["p"])
		return scrypt.Key(authArray, salt, n, r, p, dkLen)

	} else if cryptoJSON.KDF == "pbkdf2" {
		c := ensureInt(cryptoJSON.KDFParams["c"])
		prf := cryptoJSON.KDFParams["prf"].(string)
		if prf != "hmac-sha256" {
			return nil, fmt.Errorf("Unsupported PBKDF2 PRF: %s", prf)
		}
		key := pbkdf2.Key(authArray, salt, c, dkLen, sha256.New)
		return key, nil
	}

	return nil, fmt.Errorf("Unsupported KDF: %s", cryptoJSON.KDF)
}

// TODO: can we do without this when unmarshalling dynamic JSON?
// why do integers in KDF params end up as float64 and not int after
// unmarshal?
func ensureInt(x interface{}) int {
	res, ok := x.(int)
	if !ok {
		res = int(x.(float64))
	}
	return res
}
