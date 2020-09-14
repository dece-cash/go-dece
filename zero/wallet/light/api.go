package light

import (
	"bytes"
	"sort"

	"github.com/dece-cash/go-dece/czero/c_type"
	"github.com/dece-cash/go-dece/log"
	"github.com/dece-cash/go-dece/rlp"
	"github.com/dece-cash/go-dece/zero/txtool"
)

var Current_light *LightNode

func (self *LightNode) CurrentLight() *LightNode {
	return Current_light
}

func (self *LightNode) GetOutsByPKr(pkrs []c_type.PKr, start, end uint64) (br BlockOutResp, e error) {
	br.CurrentNum = self.getLastNumber()
	blockOuts := []BlockOut{}
	if end == 0 {
		end = br.CurrentNum
	}
	for _, pkr := range pkrs {
		//uPKr := pkr.ToUint512()
		prefix := append(pkrPrefix, pkr[:]...)
		iterator := self.db.NewIteratorWithPrefix(prefix)

		for ok := iterator.Seek(pkrKey(pkr, start)); ok; ok = iterator.Next() {

			key := iterator.Key()
			num := bytesToUint64(key[99:107])
			if num > end {
				break
			}
			var bds []BlockData
			if err := rlp.Decode(bytes.NewReader(iterator.Value()), &bds); err != nil {
				log.Error("Light Invalid block RLP", "Num:", num, "err:", err)
				return br, err
			} else {
				blockOut := BlockOut{Num: num, Data: bds}
				blockOuts = append(blockOuts, blockOut)
			}
		}
	}
	br.BlockOuts = blockOuts
	return br, nil
}

func (self *LightNode) GetPendingOuts(pkrs []c_type.PKr) (br BlockOutResp, e error) {
	blockOuts := []BlockOut{}

	numBlokcDatas := self.CurrentLight().getImmatureTx(pkrs)

	if pendingBlockOuts, ok := numBlokcDatas[0]; ok {
		if len(pendingBlockOuts) > 0 {
			blockOut := BlockOut{Num: 0, Data: pendingBlockOuts}
			blockOuts = append(blockOuts, blockOut)
		}
	}

	immatureBlokOuts := BlocOuts{}
	for k, v := range numBlokcDatas {
		if k != 0 {
			blockOut := BlockOut{Num: k, Data: v}
			immatureBlokOuts = append(immatureBlokOuts, blockOut)
		}

	}
	sort.Sort(immatureBlokOuts)
	blockOuts = append(blockOuts, immatureBlokOuts[:]...)
	br.BlockOuts = blockOuts
	return br, nil
}

func (self *LightNode) CheckNil(Nils []c_type.Uint256) (nilResps []NilValue, e error) {
	if len(Nils) == 0 {
		return
	}
	for _, Nil := range Nils {
		if data, err := self.db.Get(nilKey(Nil)); err != nil {
			continue
		} else {

			nilResp := NilValue{}
			if err := rlp.DecodeBytes(data, &nilResp); err != nil {
				continue
			} else {
				nilResp.Nil = Nil
				nilResps = append(nilResps, nilResp)
			}
		}
	}
	return nilResps, nil
}

type BlockOutResp struct {
	CurrentNum uint64
	BlockOuts  []BlockOut
}

type BlockOut struct {
	Num  uint64
	Data []BlockData
}

type BlockData struct {
	TxInfo TxInfo
	Out    txtool.Out
}

type BlocOuts []BlockOut

func (s BlocOuts) Len() int           { return len(s) }
func (s BlocOuts) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s BlocOuts) Less(i, j int) bool { return s[i].Num > s[j].Num }
