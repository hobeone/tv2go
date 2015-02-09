package db

import "testing"

func TestDB(t *testing.T) {
	dbh := NewMemoryDBHandle(true, true)

	a := AnimeShow{
		Name: "Yowamushi Pedal",
	}
	dbh.db.Create(&a)
}
