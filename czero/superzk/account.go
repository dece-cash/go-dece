package superzk

import "C"
import (
	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/c_type"
)

func Seed2Sk(seed *c_type.Uint256) (ret c_type.Uint512) {
	return c_superzk.Seed2Sk(seed)
}

func Sk2Tk(sk *c_type.Uint512) (tk c_type.Tk, e error) {
	return c_superzk.Sk2Tk(sk)
}

func Tk2Pk(tk *c_type.Tk) (ret c_type.Uint512, e error) {
	return c_superzk.Tk2Pk(tk)
}

func Pk2PKr(addr *c_type.Uint512, r *c_type.Uint256) (pkr c_type.PKr) {
	pkr, _ = c_superzk.Pk2PKr(addr, r)
	return
}

func IsPKValid(pk *c_type.Uint512) bool {
	return c_superzk.IsPKValid(pk)
}

func IsPKrValid(pkr *c_type.PKr) bool {
	return c_superzk.IsPKrValid(pkr)
}

func IsMyPKr(tk *c_type.Tk, pkr *c_type.PKr) (succ bool) {
	return c_superzk.IsMyPKr(tk, pkr)
}

func SignPKr_ByHeight(num uint64, sk *c_type.Uint512, data *c_type.Uint256, pkr *c_type.PKr) (sign c_type.Uint512, e error) {
	return c_superzk.SignPKr_X(sk, data, pkr)
}

func VerifyPKr_ByHeight(num uint64, data *c_type.Uint256, sign *c_type.Uint512, pkr *c_type.PKr) bool {
	return c_superzk.VerifyPKr_X(data, sign, pkr)
}

func FetchRootCM(tk *c_type.Tk, nl *c_type.Uint256, baser *c_type.Uint256) (root_cm c_type.Uint256, e error) {
	return c_superzk.FetchRootCM(tk, nl, baser)
}
