package generate_1

import (
	"github.com/dece-cash/go-dece/zero/txs/assets"
	"github.com/dece-cash/go-dece/zero/txs/stx/tx"
	"github.com/dece-cash/go-dece/zero/txtool"
	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/c_type"
)

func ConfirmOutC(key *c_type.Uint256, outc *tx.Out_C) (dout *txtool.TDOut, ar c_type.Uint256) {
	info := c_superzk.DecInfoDesc{}
	info.Key = *key
	info.Einfo = outc.EInfo
	c_superzk.DecOutput(&info)
	asset_desc := c_superzk.AssetDesc{}
	asset_desc.Asset = info.Asset_ret
	asset_desc.Ar = info.Ar_ret
	ar = asset_desc.Ar
	if e := c_superzk.GenAssetCM(&asset_desc); e != nil {
		return
	}
	if asset_desc.Asset_cm_ret == outc.AssetCM {
		dout = &txtool.TDOut{}
		dout.Asset = assets.NewAssetByType(&info.Asset_ret)
		dout.Memo = info.Memo_ret
		return
	} else {
		return
	}
}
