package messages

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "bitbucket.org/shareringvn/cosmos-sdk/types"

	"github.com/sharering/shareledger/constants"
	"github.com/sharering/shareledger/types"
)

type MsgCreate struct {
	FromDenom string    `json:"from_denom"`
	ToDenom   string    `json:"to_denom"`
	Rate      types.Dec `json:"rate"`
}

var _ sdk.Msg = MsgCreate{}

func NewMsgCreate(
	from string,
	to string,
	rate types.Dec,
) MsgCreate {
	return MsgCreate{
		FromDenom: from,
		ToDenom:   to,
		Rate:      rate,
	}
}

// Type type of this message
func (msg MsgCreate) Type() string {
	return constants.MESSAGE_EXCHANGE_RATE
}

func (msg MsgCreate) ValidateBasic() sdk.Error {
	if msg.FromDenom == msg.ToDenom {
		return sdk.ErrInternal(fmt.Sprintf(constants.EXC_SAME_DENOM, msg.FromDenom))
	}

	if !types.IsValidDenom(msg.FromDenom) || !types.IsValidDenom(msg.ToDenom) {
		return sdk.ErrInternal(fmt.Sprintf(constants.EXC_INVALID_DENOM,
			strings.Join(constants.ALL_DENOMS, ","),
			strings.Join([]string{msg.FromDenom, msg.ToDenom}, ",")))
	}

	if msg.Rate.IsZero() {
		return sdk.ErrInternal(fmt.Sprintf(constants.EXC_INVALID_RATE, msg.Rate.String()))
	}

	return nil
}

func (msg MsgCreate) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func (msg MsgCreate) String() string {
	return fmt.Sprintf("ExchangeRate/MsgCreate{%s}", msg.GetSignBytes())
}

func (msg MsgCreate) GetSigners() []sdk.Address {
	return []sdk.Address{}
}

func (msg MsgCreate) Tags() sdk.Tags {
	return sdk.NewTags("msg.module", []byte("exchangerate")).
		AppendTag("fromDenom", []byte(msg.FromDenom)).
		AppendTag("toDenom", []byte(msg.ToDenom))
}
