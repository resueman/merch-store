package entity

type PurchaseOperation struct {
	ItemID            int `db:"item_id"`
	CustomerAccountID int `db:"customer_account_id"`
	Quantity          int `db:"quantity"`
	TotalPrice        int `db:"total_price"`
}

type TransferOperation struct {
	SenderAccountID    int `db:"sender_account_id"`
	RecipientAccountID int `db:"recipient_account_id"`
	Amount             int `db:"amount"`
}
