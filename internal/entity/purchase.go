package entity

type Purchase struct {
	Name     string `db:"name"`
	Quantity int    `db:"quantity"`
}
