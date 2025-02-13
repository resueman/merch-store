package entity

type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Hash     string `db:"password"`
}

type CreateUserInput struct {
	Username string `db:"username"`
	Hash     string `db:"password"`
}
