package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBatch(t *testing.T) {
	r := &Robot{}
	_, err := r.CreateBatch()
	assert.Nil(t, err)
}

func TestAllAccounts(t *testing.T) {
	r := &Robot{}
	result, err := r.AllAcounts()
	assert.Nil(t, err)
	assert.Equal(t, 200, len(result))
}

func TestGetRobotById(t *testing.T) {
	r := &Robot{}
	res, err := r.GetById(5)
	assert.Nil(t, err)
	_ = res
}
