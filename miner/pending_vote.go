package miner

import (
	"sync"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/log"

	"github.com/dece-cash/go-dece/common"
	"github.com/dece-cash/go-dece/core/types"
)

type voteKey struct {
	headerNumber uint64
	posHash      common.Hash
}

type shareSet map[c_type.Uint512]types.Vote

type voteSet map[common.Hash]shareSet

func (self voteSet) copy() (ret voteSet) {
	ret = make(voteSet)
	for k0, v0 := range self {
		ss := make(shareSet)
		for k1, v1 := range v0 {
			ss[k1] = v1
		}
		ret[k0] = ss
	}
	return
}

type pendingVote struct {
	pendingVoteMu sync.RWMutex
	pendingVote   map[voteKey]voteSet
}

func newPendingVote() (ret pendingVote) {
	ret.pendingVote = make(map[voteKey]voteSet)
	return ret
}

func (self *pendingVote) add(vote *types.Vote) {
	self.pendingVoteMu.Lock()
	defer self.pendingVoteMu.Unlock()

	key := voteKey{vote.ParentNum + 1, vote.PosHash}
	log.Trace("pendingVote add vote", "poshash", vote.PosHash, "block", vote.ParentNum+1, "idx", vote.Idx, "sign", common.BytesToHash(vote.Sign[:]))
	var vs voteSet
	if _, ok := self.pendingVote[key]; !ok {
		vs = make(voteSet)
		self.pendingVote[key] = vs
	} else {
		vs = self.pendingVote[key]
	}

	var ss shareSet
	if _, ok := vs[vote.ShareId]; !ok {
		ss = make(shareSet)
		vs[vote.ShareId] = ss
	} else {
		ss = vs[vote.ShareId]
	}
	ss[vote.Sign] = *vote
}

func (self *pendingVote) deleteVotes(key voteKey, votes []types.Vote) {
	self.pendingVoteMu.Lock()
	defer self.pendingVoteMu.Unlock()

	if vs, ok := self.pendingVote[key]; ok {
		for _, vote := range votes {
			if _, ok := vs[vote.ShareId]; ok {
				delete(vs[vote.ShareId], vote.Sign)
			}
		}
	}
}

func (self *pendingVote) deleteBefore(num uint64) {
	self.pendingVoteMu.Lock()
	defer self.pendingVoteMu.Unlock()
	dels := []voteKey{}
	for k := range self.pendingVote {
		if k.headerNumber <= num {
			dels = append(dels, k)
		}
	}
	for _, del := range dels {
		delete(self.pendingVote, del)
	}
}

func (self *pendingVote) getMyPending(key voteKey) (ret voteSet) {
	self.pendingVoteMu.Lock()
	defer self.pendingVoteMu.Unlock()

	if votes, ok := self.pendingVote[key]; ok {
		ret = votes.copy()
	}
	return
}
