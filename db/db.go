package db

import (
	"brc20-trading-bot/constant"
	"fmt"

	"os"
	"sync/atomic"

	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

var master struct {
	dbx atomic.Value
}

func Master() *sqlx.DB {
	return master.dbx.Load().(*sqlx.DB)
}

func init() {
	dbuser := os.Getenv(constant.DBUSER)
	dbpassword := os.Getenv(constant.PASSWORD)
	host := os.Getenv(constant.HOST)
	port := os.Getenv(constant.PORT)
	dbName := os.Getenv(constant.Name)

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", host, port, dbName, dbuser, dbpassword)

	var dbx *sqlx.DB
	var err error
	dbx, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		panic(err)
	}
	master.dbx.Store(dbx)
}
