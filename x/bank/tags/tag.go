package tags

var (
	//Key - String type
	FromAddress    = "FromAddress"
	ToAddress      = "ToAddress"
	Amount         = "Amount"
	Event          = "Event"
	AccountAddress = "AccountAddress"

	//Value -  []byte
	Transfered = []byte("Transfered") //Transfer event fromAddress To Address
	Credit     = []byte("Credit")     //event for credit
)
