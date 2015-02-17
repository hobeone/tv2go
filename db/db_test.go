package db

import "testing"

func TestDBValidations(t *testing.T) {
	NewMemoryDBHandle(true, true)
}
