package posTypes

import (
	"time"

	// sdk "bitbucket.org/shareringvn/cosmos-sdk/types"
	"github.com/sharering/shareledger/constants"
	"github.com/sharering/shareledger/types"
)

// defaultUnbondingTime reflects three weeks in seconds as the default
// unbonding time.
//const defaultUnbondingTime time.Duration = 60 * 60 * 24 * 3 * time.Second //3 weeks -> adjust it

// Params defines the high level settings for staking
type Params struct {
	GoalBonded types.Dec `json:"goal_bonded"` // Goal of percent bonded atoms

	UnbondingTime time.Duration `json:"unbonding_time"`

	MaxValidators uint16 `json:"max_validators"` // maximum number of validators
	BondDenom     string `json:"bond_denom"`     // bondable coin denomination
}

// Equal returns a boolean determining if two Param types are identical.
/*
func (p Params) Equal(p2 Params) bool {
	bz1 := MsgCdc.MustMarshalBinary(&p)
	bz2 := MsgCdc.MustMarshalBinary(&p2)
	return bytes.Equal(bz1, bz2)
}
*/

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		GoalBonded:    types.ZeroDec(),
		UnbondingTime: constants.UNBONDING_TIME,
		MaxValidators: 10,
		BondDenom:     constants.POS_DENOM,
	}
}
