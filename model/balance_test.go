package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBRC20TokenBalanceFromDB(t *testing.T) {
	result, err := NewBRC20TokenBalanceFromDB("test", "0x01")
	assert.Nil(t, err)
	_ = result
}

func Test_BalanceInsertToDB(t *testing.T) {
	balance := &BRC20TokenBalance{
		Ticker:         "test",
		Address:        "0x01",
		OverallBalance: "100000",
	}
	err := balance.InsertToDB()
	assert.Nil(t, err)
}

func Test_BalanceUpdateToDB(t *testing.T) {
	balance := &BRC20TokenBalance{
		Ticker:         "test",
		Address:        "0x01",
		OverallBalance: "200000",
	}
	err := balance.UpdateToDB()
	assert.Nil(t, err)
}

func TestGetTickerHoldersMap(t *testing.T) {
	b := &BRC20TokenBalance{}
	result, err := b.GetTickerHoldersMap([]string{"test"})
	assert.Nil(t, err)
	_ = result
}
