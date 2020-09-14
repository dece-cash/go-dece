package utils

import (
	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/c_type"
)

type pkrsChecker struct {
	isSzk bool
}

func NewPKrChecker() (ret pkrsChecker) {
	return
}

func (self *pkrsChecker) AddPKr(pkr *c_type.PKr) {
	if c_superzk.IsSzkPKr(pkr) {
		self.isSzk = true
	}
}

func (self *pkrsChecker) AddPK(pk *c_type.Uint512) {
	if c_superzk.IsSzkPK(pk) {
		self.isSzk = true
	}
}

func (self *pkrsChecker) AddNil(nl *c_type.Uint256) {
	if c_superzk.IsSzkNil(nl) {
		self.isSzk = true
	}
}

func (self *pkrsChecker) IsSzk() (ret bool) {
	return self.isSzk
}
