// copyright 2018 The dece.cash Authors
// This file is part of the go-dece library.
//
// The go-dece library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-dece library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-dece library. If not, see <http://www.gnu.org/licenses/>.

package txstate

import (
	"fmt"
	"github.com/dece-cash/go-dece/zero/txs/stx"
	"sort"
	"sync"

	"github.com/dece-cash/go-dece/czero/c_superzk"

	"github.com/dece-cash/go-dece/zero/txs/stx/tx"

	"github.com/dece-cash/go-dece/zero/txs/zstate/merkle"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/decedb"

	"github.com/dece-cash/go-dece/zero/txs/zstate/txstate/data_v1"

	"github.com/dece-cash/go-dece/common"

	"github.com/dece-cash/go-dece/core/types"

	"github.com/pkg/errors"

	"github.com/dece-cash/go-dece/zero/txs/zstate/txstate/data"

	"github.com/dece-cash/go-dece/zero/localdb"

	"github.com/dece-cash/go-dece/zero/txs/zstate/tri"
)

type State struct {
	tri       tri.Tri
	rw        *sync.RWMutex
	num       uint64
	CzeroTree merkle.MerkleTree
	SzkTree   merkle.MerkleTree

	data data.IData

	logs      []data.Log
	revisions []data.Revision
}

func (self *State) Num() uint64 {
	return self.num
}

func (self *State) Tri() tri.Tri {
	return self.tri
}

func NewState(tri tri.Tri, num uint64) (state State) {
	state = State{tri: tri, num: num}
	state.rw = new(sync.RWMutex)
	state.CzeroTree = CzeroMerkleParam.NewMerkleTree(tri)
	state.SzkTree = SzkMerkleParam.NewMerkleTree(tri)
	state.data = data_v1.NewData(num)
	state.data.Clear()
	state.load()
	return
}

func (self *State) RecordState(putter decedb.Putter, root *c_type.Uint256) {
	self.data.RecordState(putter, root)
}

func (self *State) load() {
	self.data.LoadState(self.tri)
}

func (self *State) Update() {
	self.rw.Lock()
	defer self.rw.Unlock()
	self.data.SaveState(self.tri)
	return
}

func (state *State) Snapshot(revid int) {
	state.revisions = append(state.revisions, data.Revision{revid, len(state.logs)})
}

func (state *State) Revert(revid int) {

	idx := sort.Search(len(state.revisions), func(i int) bool {
		return state.revisions[i].Id >= revid
	})
	if idx == len(state.revisions) || state.revisions[idx].Id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}

	index := state.revisions[idx].JournalIndex

	state.revisions = state.revisions[:idx]
	state.logs = state.logs[:index]

	state.data.Clear()
	for _, log := range state.logs {
		log.Op(state.data)
	}
}

func (self *State) addOut_Log(root *c_type.Uint256, out *localdb.OutState, txhash *c_type.Uint256) {
	clone := out.Clone()
	if txhash != nil {
		self.logs = append(self.logs, data.AddOutLog{root.NewRef(), &clone, txhash.NewRef()})
	} else {
		self.logs = append(self.logs, data.AddOutLog{root.NewRef(), &clone, nil})
	}

	self.data.AddOut(root, out, txhash)
	return
}
func (self *State) addNil_Log(in *c_type.Uint256) {
	self.logs = append(self.logs, data.AddNilLog{in.NewRef()})
	self.data.AddNil(in)
}
func (self *State) addDel_Log(in *c_type.Uint256) {
	self.logs = append(self.logs, data.AddDelLog{in.NewRef()})
	self.data.AddDel(in)
}

func (self *State) AddTxOut_Log(pkr *c_type.PKr) int {
	self.logs = append(self.logs, data.AddTxOutLog{pkr.NewRef()})
	return self.data.AddTxOut(pkr)
}

func (state *State) AddOut_P(out_p *tx.Out_P, txhash *c_type.Uint256) (root c_type.Uint256) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.addOut_P(out_p, txhash)
}

func (state *State) insertOS(os *localdb.OutState, txhash *c_type.Uint256) (root c_type.Uint256) {
	{
		os.Index = state.SzkTree.GetLeafSize()
		os.GenRootCM()
		root = state.SzkTree.AppendLeaf(*os.RootCM)
		state.addOut_Log(&root, os, txhash)
	}
	return
}

func (state *State) addOut_C(out_c *tx.Out_C, txhash *c_type.Uint256) (root c_type.Uint256) {
	os := localdb.OutState{}
	if out_c != nil {
		o := out_c.Clone()
		os.Out_C = &o
	}
	return state.insertOS(&os, txhash)
}

func (state *State) addOut_P(out_p *tx.Out_P, txhash *c_type.Uint256) (root c_type.Uint256) {
	os := localdb.OutState{}
	if out_p != nil {
		o := out_p.Clone()
		os.Out_P = &o
	}
	return state.insertOS(&os, txhash)
}

func (self *State) HasIn(hash *c_type.Uint256) (exists bool) {
	self.rw.Lock()
	defer self.rw.Unlock()
	return self.data.HasIn(self.tri, hash)
}

func (state *State) addTx1(tx *tx.Tx, txhash *c_type.Uint256) (e error) {

	for _, in := range tx.Ins_P {
		if !state.data.HasIn(state.tri, &in.Nil) {
			if !state.data.HasIn(state.tri, &in.Root) {
				state.addNil_Log(&in.Nil)
				state.addNil_Log(&in.Root)
			} else {
				e = errors.New("tx1.in_p.root already be used !")
				return
			}
		} else {
			e = errors.New("tx1.in_p.nil already be used !")
			return
		}
	}
	for _, in := range tx.Ins_C {
		if !state.data.HasIn(state.tri, &in.Nil) {
			state.addNil_Log(&in.Nil)
		} else {
			e = errors.New("tx1.in_c.nil already be used !")
			return
		}
	}
	for _, out := range tx.Outs_C {
		state.addOut_C(&out, txhash)
	}
	for _, out := range tx.Outs_P {
		if c_superzk.IsSzkPKr(&out.PKr) {
			state.addOut_P(&out, txhash)
		}
	}
	return
}

func (state *State) AddStx(st *stx.T) (e error) {
	state.rw.Lock()
	defer state.rw.Unlock()
	txhash := st.ToHash()

	if st.Tx1.Count() > 0 {
		return state.addTx1(&st.Tx1, &txhash)
	}

	return
}

func (state *State) GetOut(root *c_type.Uint256) (src *localdb.OutState) {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.data.GetOut(state.tri, root)
}

func (state *State) FindAnchorInSzk(root *c_type.Uint256) bool {
	state.rw.Lock()
	defer state.rw.Unlock()
	return state.data.HashRoot(state.tri, root)
}

func (self *State) GetBlockRoots() (roots []c_type.Uint256) {
	return self.data.GetRoots()
}

func (self *State) GetBlockDels() (dels []c_type.Uint256) {
	return self.data.GetDels()
}

type Chain interface {
	GetBlock(hash common.Hash, number uint64) *types.Block
}
