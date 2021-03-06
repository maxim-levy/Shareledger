package posTypes

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "bitbucket.org/shareringvn/cosmos-sdk/types"
	"bitbucket.org/shareringvn/cosmos-sdk/wire"
	"github.com/sharering/shareledger/types"
)

var MaxPartialToken types.Dec = types.NewDec(2) //100/2

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// validator, the validator is credited with a Delegation whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
type Validator struct {
	Owner   sdk.Address      `json:"owner"`   // sender of BondTx - UnbondTx returns here
	PubKey  types.PubKey     `json:"pub_key"` // pubkey of validator
	Revoked bool             `json:"revoked"` // has the validator  been revoked from bonded status?
	Status  types.BondStatus `json:"status"`  // validator status (bonded/unbonding/unbonded)

	Tokens          types.Dec `json:"tokens"`           // delegated tokens (incl. self-delegation)
	DelegatorShares types.Dec `json:"delegator_shares"` // total shares issued to a validator's delegators
	CommissionRate  types.Dec `json:"commission_rate"`  // commision kept by this validator

	Description        Description `json:"description"`           // description terms for the validator
	BondHeight         int64       `json:"bond_height"`           // earliest height as a bonded validator
	BondIntraTxCounter int16       `json:"bond_intra_tx_counter"` // block-local tx index of validator change
	UnbondingHeight    int64       `json:"unbonding_height"`      // if unbonding, height at which this validator has begun unbonding
	UnbondingMinTime   int64       `json:"unbonding_time"`        // time.Time  // if unbonding, min time for the validator to complete unbonding
	//	ProposerRewardPool sdk.Coins   `json:"proposer_reward_pool"`  // XXX reward pool collected from being the proposer

}

// enforce the Validator type at compile time
var _ types.Validator = Validator{}

// Validators - list of Validators
type Validators []Validator

// to encode/decode of Validator
type validatorValue struct {
	PubKey             types.PubKey
	Revoked            bool
	Status             types.BondStatus
	Tokens             types.Dec
	DelegatorShares    types.Dec
	CommissionRate     types.Dec
	Description        Description
	BondHeight         int64
	BondIntraTxCounter int16
	UnbondingHeight    int64
	UnbondingMinTime   int64 //time.Time
}

// NewValidator - initialize a new validator
func NewValidator(owner sdk.Address, pubKey types.PubKey, description Description) Validator {
	return Validator{
		Owner:   owner,
		PubKey:  pubKey,
		Revoked: false,
		Status:  types.Unbonded,

		Tokens:             types.ZeroDec(),
		DelegatorShares:    types.OneDec(),
		CommissionRate:     types.ZeroDec(),
		Description:        description,
		BondHeight:         int64(0),
		BondIntraTxCounter: int16(0),
		UnbondingHeight:    int64(0),
		UnbondingMinTime:   int64(0), //time.Unix(0, 0).UTC(),
		//ProposerRewardPool: sdk.Coins{},
	}
}

// only the vitals - does not check bond height of IntraTxCounter
func (v Validator) Equal(c2 Validator) bool {
	return v.PubKey.Equals(c2.PubKey) &&
		bytes.Equal(v.Owner, c2.Owner) &&
		// v.PoolShares.Equal(c2.PoolShares) &&
		v.Tokens.Equal(c2.Tokens) &&
		v.DelegatorShares.Equal(c2.DelegatorShares) &&
		v.Description == c2.Description //&&
	//v.BondHeight == c2.BondHeight &&
	//v.BondIntraTxCounter == c2.BondIntraTxCounter && // counter is always changing
	// v.ProposerRewardPool.IsEqual(c2.ProposerRewardPool)
}

const DoNotModifyDes = "[do-not-modify]"

// Description - description fields for a validator
type Description struct {
	Moniker  string `json:"moniker"`  // name
	Identity string `json:"identity"` // optional identity signature (ex. UPort or Keybase)
	Website  string `json:"website"`  // optional website link
	Details  string `json:"details"`  // optional details
}

func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

func (v Validator) GetABCIPubKey() crypto.PubKeySecp256k1 {
	if pk, ok := v.GetPubKey().(types.PubKeySecp256k1); ok {
		return pk.ToABCIPubKey()
	} else {
		panic("PubKey is not of PubKeySecp256k1")
	}

}

// validator which fulfills abci validator interface for use in Tendermint
// ABCIValidator returns an abci.Validator from a staked validator type.
func (v Validator) ABCIValidator() abci.Validator {
	return abci.Validator{
		PubKey:  tmtypes.TM2PB.PubKey(v.GetABCIPubKey()),
		Address: v.GetABCIPubKey().Address(),
		// Address: v.GetPubKey().Address(),
		Power: v.GetPower().RoundInt64(), //v.BondedTokens().RoundInt64(),
	}
}

// ABCIValidator returns an abci.Validator from a staked validator type.
func (v Validator) ABCIValidatorZero() abci.Validator {
	return abci.Validator{
		PubKey:  tmtypes.TM2PB.PubKey(v.GetABCIPubKey()),
		Address: v.GetABCIPubKey().Address(),
		// Address: v.GetPubKey().Address(),
		Power: 0,
	}
}

// UpdateStatus updates the location of the shares within a validator
// to reflect the new status
func (v Validator) UpdateStatus(pool Pool, NewStatus types.BondStatus) (Validator, Pool) {

	switch v.Status {

	case types.Unbonded:

		switch NewStatus {
		case types.Unbonded:
			return v, pool
		case types.Bonded:
			pool = pool.looseTokensToBonded(v.Tokens)
		}
	case types.Unbonding:

		switch NewStatus {
		case types.Unbonding:
			return v, pool
		case types.Bonded:
			pool = pool.looseTokensToBonded(v.Tokens)
		}
	case types.Bonded:

		switch NewStatus {
		case types.Bonded:
			return v, pool
		default:
			pool = pool.bondedTokensToLoose(v.Tokens)
		}
	}

	v.Status = NewStatus
	return v, pool
}

// Returns if the validator should be considered unbonded
func (v Validator) IsUnbonded(ctx sdk.Context) bool {
	switch v.Status {
	case types.Unbonded:
		return true
	case types.Unbonding:
		//todo: check the time if it surpass the unboundtingTime
		//ctxTime := ctx.BlockHeader().Time

		//if ctxTime.After(v.UnbondingMinTime) {
		//		return true
		//	}
		return false
	}
	return false
}

// removes tokens from a validator
func (v Validator) RemoveTokens(pool Pool, tokens types.Dec) (Validator, Pool) {
	if v.Status == types.Bonded {
		pool = pool.bondedTokensToLoose(tokens)
	}

	v.Tokens = v.Tokens.Sub(tokens)
	return v, pool
}

// SetInitialCommission attempts to set a validator's initial commission. An
// error is returned if the commission is invalid.
// func (v Validator) SetInitialCommission(commission Commission) (Validator, sdk.Error) {
// 	if err := commission.Validate(); err != nil {
// 		return v, err
// 	}

// 	v.Commission = commission
// 	return v, nil
// }

//_________________________________________________________________________________________________________

// XXX Audit this function further to make sure it's correct
// add tokens to a validator
func (v Validator) AddTokensFromDel(pool Pool, amount types.Dec) (Validator, Pool, types.Dec) {

	// bondedShare/delegatedShare
	exRate := v.DelegatorShareExRate()
	amountDec := amount

	if v.Status == types.Bonded {
		pool = pool.looseTokensToBonded(amountDec)
	}

	v.Tokens = v.Tokens.Add(amountDec)
	issuedShares := amountDec.Quo(exRate)
	v.DelegatorShares = v.DelegatorShares.Add(issuedShares)

	return v, pool, issuedShares
}

// RemoveDelShares removes delegator shares from a validator.
func (v Validator) RemoveDelShares(pool Pool, delShares types.Dec) (Validator, Pool, types.Dec) {
	issuedTokens := v.DelegatorShareExRate().Mul(delShares)
	v.Tokens = v.Tokens.Sub(issuedTokens)
	v.DelegatorShares = v.DelegatorShares.Sub(delShares)

	if v.Status == types.Bonded {
		pool = pool.bondedTokensToLoose(issuedTokens)
	}

	return v, pool, issuedTokens
}

// DelegatorShareExRate gets the exchange rate of tokens over delegator shares.
// UNITS: tokens/delegator-shares
func (v Validator) DelegatorShareExRate() types.Dec {
	if v.DelegatorShares.IsZero() || v.Tokens.IsZero() {
		return types.OneDec()
	}
	return v.Tokens.Quo(v.DelegatorShares)
}

// Get the bonded tokens which the validator holds
func (v Validator) BondedTokens() types.Dec {
	if v.Status == types.Bonded {
		return v.Tokens
	}
	return types.ZeroDec()
}

//check if the adding token violate the percent rule or not:
func (v Validator) IsDelegatingTokenValid(pool Pool, tokenAMount types.Dec) bool {
	totalValToken := v.Tokens.Add(tokenAMount)
	return totalValToken.LT(pool.TokenSupply().Quo(MaxPartialToken))
}

// unmarshal a redelegation from a store key and value
func UnmarshalValidator(cdc *wire.Codec, owner sdk.Address, value []byte) (validator Validator, err error) {
	//TODO: Checking owner address
	/*
		if len(owner) != types.ADDRESSLENGTH {
			err = fmt.Errorf("%v", err.ErrBadValidatorAddr(DefaultCodespace).Data())
			return
		}*/

	var storeValue validatorValue
	err = cdc.UnmarshalBinary(value, &storeValue)
	if err != nil {
		return
	}

	return Validator{
		Owner:              owner,
		PubKey:             storeValue.PubKey,
		Revoked:            storeValue.Revoked,
		Tokens:             storeValue.Tokens,
		Status:             storeValue.Status,
		DelegatorShares:    storeValue.DelegatorShares,
		CommissionRate:     storeValue.CommissionRate,
		Description:        storeValue.Description,
		BondHeight:         storeValue.BondHeight,
		BondIntraTxCounter: storeValue.BondIntraTxCounter,
		UnbondingHeight:    storeValue.UnbondingHeight,
		UnbondingMinTime:   storeValue.UnbondingMinTime,
	}, nil
}

// unmarshal a redelegation from a store key and value
func MustUnmarshalValidator(cdc *wire.Codec, operatorAddr, value []byte) Validator {
	validator, err := UnmarshalValidator(cdc, operatorAddr, value)
	if err != nil {
		panic(err)
	}
	return validator
}

// return the redelegation without fields contained within the key for the store
func MustMarshalValidator(cdc *wire.Codec, validator Validator) []byte {
	val := validatorValue{
		PubKey:             validator.PubKey,
		Revoked:            validator.Revoked,
		Status:             validator.Status,
		Tokens:             validator.Tokens,
		DelegatorShares:    validator.DelegatorShares,
		CommissionRate:     validator.CommissionRate,
		Description:        validator.Description,
		BondHeight:         validator.BondHeight,
		BondIntraTxCounter: validator.BondIntraTxCounter,
		UnbondingHeight:    validator.UnbondingHeight,
		UnbondingMinTime:   validator.UnbondingMinTime,
	}
	return cdc.MustMarshalBinary(val)
}

//______________________________________________________________________

var _ types.Validator = Validator{}

// constant used in flags to indicate that description field should not be updated
const DoNotModifyDesc = "[do-not-modify]"

// nolint - for sdk.Validator
func (v Validator) GetMoniker() string          { return v.Description.Moniker }
func (v Validator) GetStatus() types.BondStatus { return v.Status }

func (v Validator) GetOwner() sdk.Address   { return v.Owner }
func (v Validator) GetPubKey() types.PubKey { return v.PubKey }
func (v Validator) GetPower() types.Dec {

	if v.BondedTokens().IsZero() {
		// fmt.Printf("BondedTokens")
		return types.ZeroDec()
	}
	//calculate power based on Logarit
	bondedToken := v.BondedTokens().RoundInt64()
	//s := fmt.Sprintf("%v", math.Log2(float64(bondedToken)))
	power := int64(math.Log2(float64(bondedToken)))

	// fmt.Printf("Power: %d\n", power)

	if power >= 20 {
		power = power - 20
	}
	// fmt.Printf("Power1: %d\n", power)
	if power <= 0 {
		// fmt.Printf("Power2: %d\n", power)
		return types.OneDec()
	}
	return types.NewDec(power)
}

func (v Validator) GetDelegatorShares() types.Dec { return v.DelegatorShares }
func (v Validator) GetBondHeight() int64          { return v.BondHeight }

// UpdateDescription updates the fields of a given description. An error is
// returned if the resulting description contains an invalid length.
func (d Description) UpdateDescription(d2 Description) (Description, sdk.Error) {
	if d2.Moniker == DoNotModifyDesc {
		d2.Moniker = d.Moniker
	}
	if d2.Identity == DoNotModifyDesc {
		d2.Identity = d.Identity
	}
	if d2.Website == DoNotModifyDesc {
		d2.Website = d.Website
	}
	if d2.Details == DoNotModifyDesc {
		d2.Details = d.Details
	}

	return Description{
		Moniker:  d2.Moniker,
		Identity: d2.Identity,
		Website:  d2.Website,
		Details:  d2.Details,
	}.EnsureLength()
}

// EnsureLength ensures the length of a validator's description.
func (d Description) EnsureLength() (Description, sdk.Error) {
	if len(d.Moniker) > 70 {
		return d, ErrDescriptionLength(DefaultCodespace, "moniker", len(d.Moniker), 70)
	}
	if len(d.Identity) > 3000 {
		return d, ErrDescriptionLength(DefaultCodespace, "identity", len(d.Identity), 3000)
	}
	if len(d.Website) > 140 {
		return d, ErrDescriptionLength(DefaultCodespace, "website", len(d.Website), 140)
	}
	if len(d.Details) > 280 {
		return d, ErrDescriptionLength(DefaultCodespace, "details", len(d.Details), 280)
	}

	return d, nil
}

//Human Friendly pretty printer
func (v Validator) HumanReadableString() (string, error) {

	resp := "Validator \n"
	resp += fmt.Sprintf("Owner: %s\n", v.Owner.String())
	resp += fmt.Sprintf("Validator: %s\n", v.PubKey.String())
	//resp += fmt.Sprintf("Shares: Status %s,  Amount: %s\n", sdk.BondStatusToString(v.PoolShares.Status), v.PoolShares.Amount.String())
	resp += fmt.Sprintf("Delegator Shares: %s\n", v.DelegatorShares.String())
	resp += fmt.Sprintf("Commission Rate: %s\n", v.CommissionRate.String())
	resp += fmt.Sprintf("Description: %s\n", v.Description)
	resp += fmt.Sprintf("Bond Height: %d\n", v.BondHeight)
	//	resp += fmt.Sprintf("Proposer Reward Pool: %s\n", v.ProposerRewardPool.String())

	return resp, nil
}

//-----------------------------------------------------------

// Sort abci.Validator By Address

type SortValidators []abci.Validator

func (av SortValidators) Len() int           { return len(av) }
func (av SortValidators) Swap(i, j int)      { av[i], av[j] = av[j], av[i] }
func (av SortValidators) Less(i, j int) bool { return bytes.Compare(av[i].Address, av[j].Address) < 0 }

// SortABCIValidators - function to sort abci.Validator in ascending order before returning to Tendermint
func SortABCIValidators(av []abci.Validator) []abci.Validator {
	sort.Sort(SortValidators(av))
	return av
}
