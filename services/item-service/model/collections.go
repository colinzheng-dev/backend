package model

// ItemCollection is a row in the item_colls table.
type ItemCollection struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Owner string `db:"owner"`
}
