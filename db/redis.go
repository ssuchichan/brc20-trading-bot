package db

import (
	"brc20-trading-bot/constant"
	"os"
	"sync/atomic"

	libredis "github.com/redis/go-redis/v9"
)

var mredis struct {
	rdb atomic.Value
}

func MRedis() *libredis.Client {
	return mredis.rdb.Load().(*libredis.Client)
}

func init() {
	redisUri := os.Getenv(constant.RedisUrl)
	opts, err := libredis.ParseURL(redisUri)
	if err != nil {
		panic(err)
	}

	r := libredis.NewClient(opts)

	mredis.rdb.Store(r)
}
