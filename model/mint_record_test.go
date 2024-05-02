package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountTotal(t *testing.T) {
	m := &MintRecord{
		Ticker: "test",
		User:   "0x01",
		Amount: "1113",
	}

	err := m.InsertToDB()
	assert.Nil(t, err)
}

func TestSum(t *testing.T) {
	m := &MintRecord{}
	result, err := m.MintTickerTotal("test")
	assert.Nil(t, err)
	fmt.Println(result)
}

func TestFindOrderList(t *testing.T) {
	m := &ListRecord{}
	res, err := m.CountOrderByTickerAndUser(UserTickerListFindParams{
		User:   "test",
		Ticker: "test",
	})
	assert.Nil(t, err)
	fmt.Println(res)
}
