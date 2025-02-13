package model

type OutgoingTransfer struct {
	Amount            int
	RecipientUsername string
}

type IncomingTransfer struct {
	Amount         int
	SenderUsername string
}
