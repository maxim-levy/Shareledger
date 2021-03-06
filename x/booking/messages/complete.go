package messages

import (
	"encoding/json"
	"fmt"

	sdk "bitbucket.org/shareringvn/cosmos-sdk/types"
	"github.com/sharering/shareledger/constants"
	tags "github.com/sharering/shareledger/x/booking/tags"
)

type MsgComplete struct {
	BookingID string `json:"bookingId"`
}

var _ sdk.Msg = MsgComplete{}

func NewMsgComplete(bookingId string) MsgComplete {
	return MsgComplete{
		BookingID: bookingId,
	}
}

func (msg MsgComplete) Type() string {
	return constants.MESSAGE_BOOKING
}

func (msg MsgComplete) ValidateBasic() sdk.Error {
	//if len(msg.Renter) == 0 {
	//return sdk.ErrInvalidAddress("Invalid address")
	//}

	return nil
}

func (msg MsgComplete) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return b
}

func (msg MsgComplete) Get(key interface{}) (value interface{}) { return nil }

func (msg MsgComplete) String() string {
	return fmt.Sprintf("Booking/MsgComplete{BookingID: %s}", msg.BookingID)
}

func (msg MsgComplete) GetSigners() []sdk.Address {
	//return []sdk.Address{msg.Renter}
	return []sdk.Address{}
}

func (msg MsgComplete) Tags() sdk.Tags {
	return sdk.NewTags(tags.Event, tags.BookingCompleted).
		AppendTag(tags.BookingId, []byte(msg.BookingID))
}
