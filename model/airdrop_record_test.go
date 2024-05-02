package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	airdrop := &AirdropRecord{
		FromUser: "test_from",
		ToUser:   "test_to",
		Amount:   "2.3",
	}
	err := airdrop.InsertToDB()
	assert.Nil(t, err)
}
