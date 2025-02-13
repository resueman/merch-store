package entity

type Transfer struct {
	Amount             int `db:"amount"`
	SenderAccountID    int `db:"sender_account_id"`
	RecipientAccountID int `db:"recipient_account_id"`
}
