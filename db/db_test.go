package db

import (
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	// t.Skip("data not same")
	result, err := Master().Exec("select * from token")
	assert.Nil(t, err)
	_ = result
}
