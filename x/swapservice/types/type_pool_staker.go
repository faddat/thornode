package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab.com/thorchain/bepswap/common"
)

// StakerUnit staker and their units in the pool
type StakerUnit struct {
	StakerID common.BnbAddress `json:"staker_id"`
	Units    sdk.Uint          `json:"units"`
}

// PoolStaker
// {
//    "p_id":"pool-BNB",
//    "tu" : "1000",
//    "ss" : {
//        "bnbStaker-1":"100",,
//        "bnbStaker-2":"100"
//    }
//}
type PoolStaker struct {
	Ticker     common.Ticker `json:"ticker"`      // ticker
	TotalUnits sdk.Uint      `json:"total_units"` // total units in the pool
	Stakers    []StakerUnit  `json:"stakers"`     // key will be staker id , which is the address on binane chain value will be UNITS
}

// NewPoolStaker create a new instance of PoolStaker
func NewPoolStaker(ticker common.Ticker, totalUnits sdk.Uint) PoolStaker {
	return PoolStaker{
		Ticker:     ticker,
		TotalUnits: totalUnits,
		Stakers:    []StakerUnit{},
	}
}

// String return the human readable string of PoolStaker
func (ps PoolStaker) String() string {
	bs := strings.Builder{}
	bs.WriteString(fmt.Sprintln("ticker: " + ps.Ticker.String()))
	bs.WriteString(fmt.Sprintf("total units: %d", ps.TotalUnits.Uint64()))
	if nil != ps.Stakers {
		for _, stakerUnit := range ps.Stakers {
			bs.WriteString(fmt.Sprintln(stakerUnit.StakerID.String() + " : " + stakerUnit.Units.String()))
		}
	}
	return bs.String()
}
func (ps *PoolStaker) GetStakerUnit(stakerID common.BnbAddress) StakerUnit {
	for _, item := range ps.Stakers {
		if item.StakerID == stakerID {
			return item
		}
	}
	return StakerUnit{
		StakerID: stakerID,
		Units:    sdk.ZeroUint(),
	}
}

// RemoveStakerUnit will remove the stakerunit with given staker id from the struct
func (ps *PoolStaker) RemoveStakerUnit(stakerID common.BnbAddress) {
	deleteIdx := -1
	for idx, item := range ps.Stakers {
		if item.StakerID == stakerID {
			deleteIdx = idx
		}
	}

	if deleteIdx != -1 {
		ps.Stakers = append(ps.Stakers[:deleteIdx], ps.Stakers[deleteIdx+1:]...)
	}
}

// UpsertStakerUnit it check whether the given staker unit is exist in the struct
// if it exist then just update it , otherwise it append it
func (ps *PoolStaker) UpsertStakerUnit(stakerUnit StakerUnit) {
	pos := -1
	for idx, item := range ps.Stakers {
		if item.StakerID == stakerUnit.StakerID {
			pos = idx
		}
	}
	if pos != -1 {
		ps.Stakers[pos] = stakerUnit
		return
	}
	ps.Stakers = append(ps.Stakers, stakerUnit)
}
