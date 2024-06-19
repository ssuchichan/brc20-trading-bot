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

func TestGetRobotById(t *testing.T) {
	r := &Robot{}
	res, err := r.GetById(5)
	assert.Nil(t, err)
	_ = res
}
