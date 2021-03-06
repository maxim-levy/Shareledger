package types

import (
	//"encoding/json"
	"fmt"

	sdk "bitbucket.org/shareringvn/cosmos-sdk/types"
	"bitbucket.org/shareringvn/cosmos-sdk/wire"
)

//-----------------------------------------------------------------
// Tx Interface
type SHRTx interface {
	sdk.Tx
	GetMsg() sdk.Msg
	GetSignature() SHRSignature
	VerifySignature() bool
	GetSignBytes() []byte
}

//------------------------------------------------------------------
// Tx

var _ SHRTx = BasicTx{}

// Simple tx to wrap the Msg.
type BasicTx struct {
	sdk.Msg   `json:"message"`
	Signature BasicSig `json:"signature"`
}

func NewBasicTx(msg sdk.Msg, sig BasicSig) BasicTx {
	return BasicTx{
		Msg:       msg,
		Signature: sig,
	}
}

// GetMsgs returns multiple messages
func (tx BasicTx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx.Msg}
}

// GetMsg returns the message of this transaction
func (tx BasicTx) GetMsg() sdk.Msg {
	return tx.Msg
}

// GetSignature returns the signature with this transaction
func (tx BasicTx) GetSignature() SHRSignature {
	return tx.Signature
}

// GetSignBytes returns Bytes to be signed
func (tx BasicTx) GetSignBytes() []byte {
	return tx.Msg.GetSignBytes()
}

// VerifySignature to verify signature
func (tx BasicTx) VerifySignature() bool {
	msg := tx.GetSignBytes()
	return tx.Signature.Verify(msg)
}

// JSON decode MsgSend.
func GetTxDecoder(cdc *wire.Codec) func([]byte) (sdk.Tx, sdk.Error) {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = BasicTx{}

		//fmt.Println("TxDecoder:", txBytes)
		//err := json.Unmarshal(txBytes, &tx)
		err := cdc.UnmarshalJSON(txBytes, &tx)

		if err != nil {
			return nil, sdk.ErrTxDecode(err.Error())
		}

		isVerified := tx.VerifySignature()
		if !isVerified {
			return nil, sdk.ErrTxDecode("InvalidSignature")
		}
		return tx, nil
	}
}

//------------------------------------------------------------------
// Signature

type SHRSignature interface {
	String() string
	Verify([]byte) bool
}

var _ SHRSignature = BasicSig{}

// Signature without Nonce
type BasicSig struct {
	PubKey    `json:"pub_key"`
	Signature `json:"signature"`
}

func NewBasicSig(key PubKey, sig Signature) BasicSig {
	return BasicSig{
		PubKey:    key,
		Signature: sig,
	}
}

func (sig BasicSig) String() string {
	return fmt.Sprintf("BaseSig{%s, %s}", sig.PubKey, sig.Signature)
}

func (sig BasicSig) Verify(msg []byte) bool {
	return sig.PubKey.VerifyBytes(msg, sig.Signature)
}
