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

package core

import (
	"errors"
	"math"
	"math/big"
	"strings"

	"github.com/dece-cash/go-dece/zero/txs/assets"
	"github.com/dece-cash/go-dece/zero/utils"

	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/core/vm"
	"github.com/dece-cash/go-dece/log"
	"github.com/dece-cash/go-dece/params"
)

var (
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

type StateTransition struct {
	gp         *GasPool
	msg        Message
	gas        uint64
	gasPrice   *big.Int
	initialGas uint64
	asset      *assets.Asset
	data       []byte
	state      vm.StateDB
	evm        *vm.EVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address

	To() *common.Address

	GasPrice() *big.Int
	Asset() *assets.Asset

	// Nonce() uint64

	Fee() assets.Token

	Data() []byte

	TxHash() common.Hash
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func IntrinsicGas(data []byte, contractCreation bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if contractCreation && len(data) > 0 {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		// Make sure we don't exceed uint64 for all data combinations
		if (math.MaxUint64-gas)/params.TxDataNonZeroGas < nz {
			return 0, vm.ErrOutOfGas
		}
		gas += nz * params.TxDataNonZeroGas

		z := uint64(len(data)) - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, vm.ErrOutOfGas
		}
		gas += z * params.TxDataZeroGas
	}
	return gas, nil
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(evm *vm.EVM, msg Message, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: msg.GasPrice(),
		asset:    msg.Asset(),
		data:     msg.Data(),
		state:    evm.StateDB,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) ([]byte, uint64, bool, error) {
	return NewStateTransition(evm, msg, gp).TransitionDb()
}

// to returns the recipient of the message.
func (st *StateTransition) to() common.Address {
	if st.msg == nil || st.msg.To() == nil /* contract creation */ {
		return common.Address{}
	}
	return *st.msg.To()
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return vm.ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) preCheck() error {
	curency := strings.ToUpper(common.BytesToString((st.msg.Fee().Currency).NewRef()[:]))
	gas := uint64(0)
	if curency != "DECE" {
		to := st.msg.To()

		tokens, tas := st.state.GetTokenRate(*to, curency)
		if tokens.Sign() == 0 || tas.Sign() == 0 {
			return errInsufficientBalanceForGas
		}
		taval := new(big.Int).Div(new(big.Int).Mul(st.msg.Fee().Value.ToRef().ToIntRef(), tas), tokens)
		if st.state.GetBalance(*to, "DECE").Cmp(taval) < 0 {
			return errInsufficientBalanceForGas
		}
		st.state.AddBalance(*to, curency, st.msg.Fee().Value.ToRef().ToRef().ToIntRef())
		st.state.SubBalance(*to, "DECE", taval)
		gas = new(big.Int).Div(taval, st.msg.GasPrice()).Uint64()
	} else {
		gas = new(big.Int).Div(st.msg.Fee().Value.ToRef().ToIntRef(), st.msg.GasPrice()).Uint64()
	}
	if err := st.gp.SubGas(gas); err != nil {
		return err
	}

	st.gas += gas
	st.initialGas = gas
	return nil
}

// TransitionDb will transition the state by applying the current message and
// returning the result including the used gas. It returns an error if failed.
// An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, usedGas uint64, failed bool, err error) {
	if err = st.preCheck(); err != nil {
		log.Error("TransitionDb", "preCheck err", err)
		return
	}
	msg := st.msg
	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil && len(st.data) > 0

	// Pay intrinsic gas
	gas, err := IntrinsicGas(st.data, contractCreation)
	if err != nil {
		return nil, 0, false, err
	}
	if err = st.useGas(gas); err != nil {
		return nil, 0, false, err
	}

	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
	)

	if contractCreation {
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, msg.Asset())
	} else {
		to := msg.To()

		if to != nil && st.state.IsContract(*to) {
			ret, st.gas, vmerr, _ = evm.Call(sender, st.to(), st.data, st.gas, msg.Asset())
		}
	}
	if vmerr != nil {
		log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, 0, false, vmerr
		}
		if msg.Asset() != nil {
			st.state.NextZState().AddTxOut(msg.From(), *msg.Asset(), msg.TxHash())
		}
	}

	st.refundGas()
	// asset := assets.Asset{Tkn: &assets.Token{
	//	Currency: *common.BytesToHash(common.LeftPadBytes([]byte("DECE"), 32)).HashToUint256(),
	//	Value:    utils.U256(*new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.gasPrice)),
	// },
	// }
	// st.state.GetZState().AddTxOut(st.evm.Coinbase, asset)

	return ret, st.gasUsed(), vmerr != nil, err
}

func (st *StateTransition) refundGas() {
	// Apply refund counter, capped to half of the used gas.
	refund := st.gasUsed() / 2
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return DECE for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)

	if remaining.Sign() > 0 {
		curency := strings.ToUpper(common.BytesToString(st.msg.Fee().Currency.NewRef()[:]))
		if curency != "DECE" {
			st.state.AddBalance(*st.msg.To(), "DECE", remaining)
			tokes, tas := st.state.GetTokenRate(*st.msg.To(), curency)
			if tokes.Sign() != 0 && tas.Sign() != 0 {
				remainToken := new(big.Int).Div(new(big.Int).Mul(remaining, tokes), tas)
				asset := assets.Asset{Tkn: &assets.Token{
					Currency: *common.BytesToHash(common.LeftPadBytes([]byte(curency), 32)).HashToUint256(),
					Value:    utils.U256(*remainToken),
				},
				}
				st.state.NextZState().AddTxOut(st.msg.From(), asset, st.msg.TxHash())
				st.state.SubBalance(*st.msg.To(), curency, remainToken)
			}
		} else {
			asset := assets.Asset{Tkn: &assets.Token{
				Currency: *common.BytesToHash(common.LeftPadBytes([]byte("DECE"), 32)).HashToUint256(),
				Value:    utils.U256(*remaining),
			},
			}
			st.state.NextZState().AddTxOut(st.msg.From(), asset, st.msg.TxHash())
		}
	}

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}
