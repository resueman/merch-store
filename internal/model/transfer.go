package model

type OutgoingTransfer struct {
	Amount             int
	RecipientAccountID int // TODO: -> recipientUsername
	RecipientUsername  string
}

type IncomingTransfer struct {
	Amount          int
	SenderAccountID int // TODO: -> senderUsername
	SenderUsername  string
}
