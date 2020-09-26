package verify_utils

import (
	"fmt"

	"github.com/dece-cash/go-dece/common/hexutil"
	"github.com/dece-cash/go-dece/log"
	"github.com/dece-cash/go-dece/zero/txs/stx"
	"github.com/dece-cash/go-dece/zero/utils"
)

func CheckUint(i *utils.U256) bool {
	return i.IsValid()
}
func ReportError(str string, tx *stx.T) (e error) {
	h := hexutil.Encode(tx.ToHash().NewRef()[:])
	log.Error("Verify Tx1 Error", "reason", str, "hash", h)
	return fmt.Errorf("Verify Tx1 Error: resean=%v , hash=%v", str, h)
}
