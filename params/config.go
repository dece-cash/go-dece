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

package params

import (
	"fmt"
	"math/big"

	"github.com/dece-cash/go-dece/common"
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash  = common.HexToHash("0x113d483242270ab0cd4ba353cc04b64a062713a5f05bbe97b8a0548e73218e70")
	AlphanetGenesisHash = common.HexToHash("0x294e3d4cc16e116c4a9e1d576f868d759b07ac6e010251eb93c55adb15a562ca")
)

var (
	// BetanetChainConfig is the chain parameters to run a node on the main network.
	BetanetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(2019),
		AutumnTwilightBlock: big.NewInt(0),
		Ethash:              new(EthashConfig),
	}

	// AlphanetChainConfig contains the chain parameters to run a node on the Ropsten test network.
	AlphanetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1000),
		AutumnTwilightBlock: big.NewInt(0),
		Ethash:              new(EthashConfig),
	}

	// RinkebyChainConfig contains the chain parameters to run a node on the Rinkeby test network.
	DevnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1024),
		AutumnTwilightBlock: big.NewInt(0),
		Ethash:              new(EthashConfig),
	}

	// AllEthashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Ethash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllEthashProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), new(EthashConfig)}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), nil}

	TestChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1),
		AutumnTwilightBlock: big.NewInt(0),
		//ConstantinopleBlock: nil,
		Ethash: new(EthashConfig),
	}
)

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	AutumnTwilightBlock *big.Int `json:"AutumnTwilightBlock,omitempty"` // AutumnTwilightBlock switch block (nil = no fork, 0 = already on AutumnTwilightBlock)

	// Various consensus engines
	Ethash *EthashConfig `json:"ethash,omitempty"`
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Ethash != nil:
		engine = c.Ethash
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v AutumnTwilight: %v Engine: %v}",
		c.ChainID,
		c.AutumnTwilightBlock,
		engine,
	)
}

// IsAutumnTwilight returns whether num is either equal to the AutumnTwilight fork block or greater.
func (c *ChainConfig) IsAutumnTwilight(num *big.Int) bool {
	return isForked(c.AutumnTwilightBlock, num)
}

//
//// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
//func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
//	return isForked(c.ConstantinopleBlock, num)
//}

// GasTable returns the gas table corresponding to the current phase (homestead or homestead reprice).
//
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) GasTable {
	return GasTableConstantinople
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	if isForkIncompatible(c.AutumnTwilightBlock, newcfg.AutumnTwilightBlock, head) {
		return newCompatError("AutumnTwilight fork block", c.AutumnTwilightBlock, newcfg.AutumnTwilightBlock)
	}
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntatic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID          *big.Int
	IsAutumnTwilight bool
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Rules{ChainID: new(big.Int).Set(chainID), IsAutumnTwilight: c.IsAutumnTwilight(num)}
}
