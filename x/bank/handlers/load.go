package handlers

import (
	"fmt"

	sdk "bitbucket.org/shareringvn/cosmos-sdk/types"

	"github.com/sharering/shareledger/utils"
	"github.com/sharering/shareledger/constants"
	"github.com/sharering/shareledger/x/auth"
	"github.com/sharering/shareledger/x/bank/messages"
)

//--------------------------------
// Handler for the message

func HandleMsgLoad(am auth.AccountMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		loadMsg, ok := msg.(messages.MsgLoad)
		if !ok {
			return sdk.NewError(2, 1, "MsgLoad is malformed").Result()
		}

		// IMPORTANT
		// TODO: require a list of limited accounts which are priviledged to load coins

		signer := auth.GetSigner(ctx)

		// Only reserve is allowed to execute this function
		if !utils.IsValidReserve(signer.GetAddress()) {
			return sdk.ErrInternal(fmt.Sprintf(constants.RES_RESERVE_ONLY)).Result()
		}

		// Credit the account
		var resT sdk.Result

		if resT = handleTo(ctx, am, loadMsg.Account, loadMsg.Amount); !resT.IsOK() {
			return resT
		}
		return sdk.Result{
			Log:  resT.Log,
			Data: resT.Data,
			Tags: loadMsg.Tags(),
		}

	}
}
