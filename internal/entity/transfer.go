package entity

type Transfer struct {
	Amount            int    `db:"amount"`
	SenderUsername    string `db:"sender_username"`
	RecipientUsername string `db:"recipient_username"`
}
