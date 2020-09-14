package stake

import (
	"math/big"

	"github.com/dece-cash/go-dece/zero/zconfig"

	"github.com/dece-cash/go-dece/czero/deceparam"
)

var (
	poolValueThreshold, _ = new(big.Int).SetString("200000000000000000000000", 10) //200K DECE
	lockingBlockNum       = uint64(198720)                                         // 1 month 30*24*60*4.6

	basePrice = big.NewInt(2000000000000000000) //2 DECE
	addition  = big.NewInt(759357240838722)

	baseReware, _ = new(big.Int).SetString("10500000000000000000", 10) // 10.5 DECE
	rewareStep    = big.NewInt(76854301391338)

	maxReware, _ = new(big.Int).SetString("35600000000000000000", 10) //35.6 DECE

	outOfDateWindow      = uint64(198720) // 1 month 30*24*60*4.6
	missVotedWindow      = uint64(596160) // 3 month 3*30*24*60*4.6
	payWindow            = uint64(42336)  // 1 week 7*24*60*4.6
	statisticsMissWindow = uint64(6048)   // 1 day 24*60*4.6

	//test
	//outOfDateWindow      = uint64(100)
	//missVotedWindow      = uint64(120)
	//payWindow            = uint64(5)
	//statisticsMissWindow = uint64(10)
)

const (
	SOLO_RATE  = 3
	TOTAL_RATE = 4

	minSharePoolSize = 20000 // 20K
	//test
	//minSharePoolSize = 20

	minMissRate    = 0.2
	MaxVoteCount   = 3
	ValidVoteCount = 2
)

func getMinSharePoolSize() uint32 {
	if zconfig.IsTestFork() {
		return 10000000
	}
	if deceparam.Is_Dev() {
		return 20
	}

	return minSharePoolSize
}

func GetPoolValueThreshold() *big.Int {
	if deceparam.Is_Dev() {
		return big.NewInt(1000000000000000000)
	}
	return poolValueThreshold
}

func GetLockingBlockNum() uint64 {
	if deceparam.Is_Dev() {
		return 10
	}
	return lockingBlockNum
}

func getStatisticsMissWindow() uint64 {
	if deceparam.Is_Dev() {
		return 10
	}
	return statisticsMissWindow
}

func getOutOfDateWindow() uint64 {
	if deceparam.Is_Dev() {
		return 100
	}
	return outOfDateWindow
}

func getMissVotedWindow() uint64 {
	if deceparam.Is_Dev() {
		return 105
	}
	return missVotedWindow
}

func getPayPeriod() uint64 {
	if deceparam.Is_Dev() {
		return 5
	}
	return payWindow
}
