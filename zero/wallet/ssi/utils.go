package ssi

import (
	"encoding/json"
	"log"

	"github.com/dece-cash/go-dece/zero/txtool/generate/generate_1"

	"github.com/dece-cash/go-dece/czero/c_superzk"
	"github.com/dece-cash/go-dece/czero/superzk"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/zero/txtool"
)

func DecNilOuts(outs []txtool.Out, skr *c_type.PKr) (douts []txtool.DOut) {
	sk := c_type.Uint512{}
	copy(sk[:], skr[:])
	tk, _ := superzk.Sk2Tk(&sk)
	for _, out := range outs {
		dout := txtool.DOut{}

		data, _ := json.Marshal(out)
		log.Printf("DecOuts out : %s", string(data))

		if out.State.OS.Out_P != nil {
			if nl, e := c_superzk.GenNil(&tk, out.State.OS.RootCM, &out.State.OS.Out_P.PKr); e == nil {
				dout.Asset = out.State.OS.Out_P.Asset.Clone()
				dout.Memo = out.State.OS.Out_P.Memo
				dout.Nil = nl
				log.Printf("DecOuts success")
			}
			log.Printf("DecOuts Out_P")
		} else if out.State.OS.Out_C != nil {
			if key, _, e := c_superzk.FetchKey(&out.State.OS.Out_C.PKr, &tk, &out.State.OS.Out_C.RPK); e == nil {
				if o, _ := generate_1.ConfirmOutC(&key, out.State.OS.Out_C); o != nil {
					if nl, e := c_superzk.GenNil(&tk, out.State.OS.RootCM.NewRef(), out.State.OS.ToPKr()); e == nil {
						dout.Asset = o.Asset
						dout.Memo = o.Memo
						dout.Nil = nl
						log.Printf("DecOuts success")
					}
				}
			}
			log.Printf("DecOuts Out_C")
		}
		douts = append(douts, dout)

		data, _ = json.Marshal(douts)
		log.Printf("DecOuts douts : %s", string(data))
	}
	return
}
