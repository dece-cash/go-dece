package verify_1

import (
	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/utils"
	"github.com/dece-cash/go-dece/zero/zconfig"
)

var verify_input_procs_pool = utils.NewProcsPool(func() int { return zconfig.G_v_thread_num })

type verify_input_desc struct {
	proof        c_type.Proof
	asset_cm_new c_type.Uint256
	zpka         c_type.Uint256
	nil          c_type.Uint256
	anchor       c_type.Uint256
}

func (self *verify_input_desc) Run() error {
	if err := c_superzk.VerifyInput(
		&self.proof,
		&self.asset_cm_new,
		&self.zpka,
		&self.nil,
		&self.anchor,
	); err != nil {
		return err
	} else {
		return nil
	}
}
