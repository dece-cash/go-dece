package ethapi

import (
	"context"
	"fmt"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/wallet/light"
)

type PublicLightNodeApi struct {
	b Backend
}

func (plna PublicLightNodeApi) GetOutsByPKr(ctx context.Context, addresses []*MixAdrress, start, end uint64) (outBlockResp light.BlockOutResp, e error) {

	pkrs := []c_type.PKr{}
	for _, address := range addresses {
		addr := *address
		if len(addr) == 96 {
			var pkr c_type.PKr
			copy(pkr[:], addr[:])
			pkrs = append(pkrs, pkr)
		} else {
			return outBlockResp, fmt.Errorf("address is invalid")
		}
	}

	return plna.b.GetOutByPKr(pkrs, start, end)
}

func (plna PublicLightNodeApi) GetPendingOuts(ctx context.Context, addresses []*PKrAddress) (outBlockResp light.BlockOutResp, e error) {

	pkrs := []c_type.PKr{}
	for _, address := range addresses {
		addr := *address
		if len(addr) == 96 {
			var pkr c_type.PKr
			copy(pkr[:], addr[:])
			pkrs = append(pkrs, pkr)
		} else {
			return outBlockResp, fmt.Errorf("address is invalid")
		}
	}
	return light.Current_light.GetPendingOuts(pkrs)

}

func (plna PublicLightNodeApi) CheckNil(Nils []c_type.Uint256) (nilResps []light.NilValue, e error) {

	return plna.b.CheckNil(Nils)
}
