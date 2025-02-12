package entity

type AccountInfo struct {
	Balance           int
	Inventory         []Inventory
	IncomingTransfers []IncomingTransfer
	OutgoingTransfers []OutgoingTransfer
}

type Inventory struct {
	Name     string `db:"name"`
	Quantity int    `db:"quantity"`
}

type OutgoingTransfer struct {
	Amount             int `db:"amount"`
	RecipientAccountID int `db:"recipient_account_id"`
}

type IncomingTransfer struct {
	Amount          int `db:"amount"`
	SenderAccountID int `db:"sender_account_id"`
}
