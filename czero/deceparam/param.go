package deceparam

func SIP1() uint64 {
	if is_dev {
		return 0
	} else {
		return uint64(140000) // for miner rewards
	}
}

func SIP2() uint64 {
	if is_dev {
		return 0
	} else {
		return uint64(200800) // for miner rewards
	}
}

func SIP3() uint64 {
	if is_dev {
		return 0
	} else {
		return uint64(200970) // for miner rewards
	}
}

func SIP4() uint64 { //WORLDSHARE booster fix
	if is_dev {
		return 0
	} else {
		return uint64(2845500) // for miner rewards
	}
}

func SIP5() uint64 { //WORLDSHARE booster fix
	if is_dev {
		return 0
	} else {
		return uint64(2863000) // for miner rewards
	}
}

func SIP6() uint64 { //WORLDSHARE booster fix
	if is_dev {
		return 0
	} else {
		return uint64(3495000) // for miner rewards
	}
}

const MAX_O_INS_LENGTH = int(2500)

const MAX_O_OUT_LENGTH = int(10)

const MAX_Z_OUT_LENGTH_OLD = int(6)

const MAX_Z_OUT_LENGTH_SIP2 = int(500)

const MAX_CONTRACT_OUT_COUNT_LENGTH = int(256)

const LOWEST_STAKING_NODE_FEE_RATE = uint32(2500)

const HIGHEST_STAKING_NODE_FEE_RATE = uint32(7500)
