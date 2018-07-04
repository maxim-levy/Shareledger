package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//------------------------------------------------------------------
// Msg

// MsgSend implements sdk.Msg
var _ sdk.Msg = MsgSend{}

// MsgSend to send coins from Input to Output
type MsgSend struct {
	Nonce  int64 `json:"nonce"`
	From   sdk.Address `json:"from"`
	To     sdk.Address `json:"to"`
	Amount Coin   `json:"amount"`
}

// NewMsgSend
func NewMsgSend(nonce int64, from, to sdk.Address, amt Coin) MsgSend {
	return MsgSend{nonce, from, to, amt}
}

// Implements Msg.
func (msg MsgSend) Type() string { return "send" }

// Implements Msg. Ensure the addresses are good and the
// amount is positive.
func (msg MsgSend) ValidateBasic() sdk.Error {
	if len(msg.From) == 0 {
		return sdk.ErrInvalidAddress("From address is empty")
	}
	if len(msg.To) == 0 {
		return sdk.ErrInvalidAddress("To address is empty")
	}
	if !msg.Amount.IsPositive() {
		return sdk.ErrInvalidCoins("Amount is not positive")
	}
	return nil
}

// Implements Msg. JSON encode the message.
func (msg MsgSend) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// Implements Msg. Return the signer.
func (msg MsgSend) GetSigners() []sdk.Address {
	return []sdk.Address{msg.From}
}

// Returns the sdk.Tags for the message
func (msg MsgSend) Tags() sdk.Tags {
	return sdk.NewTags("sender", []byte(msg.From.String())).
		AppendTag("receiver", []byte(msg.To.String()))
}
