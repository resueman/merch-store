package entity

type PurchaseOperation struct {
	ItemID            int
	CustomerAccountID int
	Quantity          int
	TotalPrice        int
}

type TransferOperation struct {
	SenderAccountID    int
	RecipientAccountID int
	Amount             int
}
