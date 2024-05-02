package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlockFromDB(t *testing.T) {
	block, err := NewBlockFromDB(1)
	assert.Nil(t, err)
	_ = block
}

func TestBlockInsert(t *testing.T) {
	b := &Block{Height: 1}
	err := b.InsertToDB()
	assert.Nil(t, err)
}

func TestFinishBlock(t *testing.T) {
	b, err := NewBlockFromDB(1)
	assert.Nil(t, err)
	err = b.FinishHandleBlock()
	assert.Nil(t, err)
}

func TestLatestHeight(t *testing.T) {
	b := &Block{}
	result, err := b.LatestHeight()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, result)
}
