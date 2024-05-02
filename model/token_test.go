package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenFromDBByTicker(t *testing.T) {
	_, err := NewTokenFromDBByTicker("test")
	assert.Nil(t, err)
}

func TestInsertToDB(t *testing.T) {
	token := &Token{
		Ticker:  "test",
		Decimal: 8,
		Max:     "100000",
		Limit:   "1000",
	}

	err := token.InsertToDB()
	assert.Nil(t, err)
}

func TestFindList(t *testing.T) {
	token := &Token{}
	_, err := token.FindPageList(1, 20, FindParams{Type: 0})
	assert.Nil(t, err)
}

func TestFindInProgresssList(t *testing.T) {
	token := &Token{}
	_, err := token.FindPageList(1, 20, FindParams{Type: 1})
	assert.Nil(t, err)
}

func TestFindCompleteList(t *testing.T) {
	token := &Token{}
	result, err := token.FindPageList(1, 20, FindParams{Type: 2})
	assert.Nil(t, err)
	assert.Equal(t, len(result.Data), 0)
}

func TestMarketIndex(t *testing.T) {
	token := &Token{}
	_, err := token.FindMarketInfos(1, 20, MarketSearchParam{Ticker: "test"})
	assert.Nil(t, err)
}
