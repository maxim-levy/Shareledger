package pos

import (
	"fmt"

	"github.com/pkg/errors"

	sdk "bitbucket.org/shareringvn/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"

	"github.com/sharering/shareledger/types"
	"github.com/sharering/shareledger/x/pos/keeper"
	posTypes "github.com/sharering/shareledger/x/pos/type"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool       posTypes.Pool        `json:"pool"`
	Params     posTypes.Params      `json:"params"`
	Validators []posTypes.Validator `json:"validators"`
	//Bonds      []Delegation `json:"bonds"`
}

func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data GenesisState) ([]abci.Validator, error) {

	var abciVals []abci.Validator
	keeper.SetPool(ctx, data.Pool)
	keeper.SetParams(ctx, data.Params)
	keeper.InitIntraTxCounter(ctx)

	for _, validator := range data.Validators {

		fmt.Printf("Validator in gensis: %v", validator)

		if validator.DelegatorShares.IsZero() {
			return abciVals, errors.Errorf("genesis validator cannot have zero delegator shares, validator: %v", validator)
		}
		abciVals = append(abciVals, validator.ABCIValidator())

		keeper.SetValidator(ctx, validator)

		vdi := posTypes.NewValidatorDistInfo(validator.Owner, int64(0))

		// Store ValidatorDistInfo
		keeper.SetValidatorDistInfo(ctx, vdi)
	}

	return abciVals, nil

}
func GenerateGenesis(pubKey types.PubKeySecp256k1) GenesisState {
	validator := posTypes.NewValidator(
		pubKey.Address(),
		pubKey,
		posTypes.NewDescription("sharering", "", "sharering.network", ""))

	gs := GenesisState{
		Pool:       posTypes.InitialPool(),
		Validators: []posTypes.Validator{validator},
	}
	return gs
}
