package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	MRedis().SetNX(context.Background(), "S:Latest:Robot", 0, time.Duration(0))
	result, err := MRedis().Get(context.Background(), "S:Latest:Robot").Int64()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), result)
}
