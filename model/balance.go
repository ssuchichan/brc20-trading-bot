package model

import (
	"brc20-trading-bot/db"
	"brc20-trading-bot/decimal"
	"database/sql"

	"time"

	"github.com/jmoiron/sqlx"
)

type BRC20TokenBalance struct {
	Base
	Address        string `json:"address" db:"address"`
	Ticker         string `json:"ticker" db:"ticker"`
	OverallBalance string `json:"overall_balance" db:"overall_balance"`
	Height         int64  `json:"height" db:"height"`
}

func NewBRC20TokenBalanceFromDB(ticker string, address string) (*BRC20TokenBalance, error) {
	// read from db
	var result BRC20TokenBalance
	err := db.Master().Get(&result, "select * from  balance where ticker = $1 and address = $2", ticker, address)
	if err != nil {
		if err == sql.ErrNoRows {
			return &BRC20TokenBalance{Ticker: ticker, Address: address}, nil
		}
		return nil, err
	}
	return &result, nil
}

func (b *BRC20TokenBalance) InsertToDB() error {
	if b.CreateTime == 0 {
		b.CreateTime = time.Now().Unix()
	}
	b.UpdateTime = time.Now().Unix()
	_, err := db.Master().NamedExec("INSERT INTO balance (address, ticker, overall_balance, create_time, update_time, height) values (:address, :ticker, :overall_balance, :create_time, :update_time, :height)", b)
	if err != nil {
		return err
	}
	return nil
}

func (b *BRC20TokenBalance) UpdateToDB() error {
	b.UpdateTime = time.Now().Unix()
	_, err := db.Master().NamedExec("update balance set overall_balance = :overall_balance, update_time = :update_time where id = :id", b)
	if err != nil {
		return err
	}
	return nil
}

func (b *BRC20TokenBalance) GetOverallBalance() (*decimal.Decimal, error) {
	if b.OverallBalance == "" {
		return decimal.NewDecimal(), nil
	}
	result, _, err := decimal.NewDecimalFromString(b.OverallBalance)
	return result, err
}

func (b *BRC20TokenBalance) GetByTickerAndAddress(ticker string, address string) (*BRC20TokenBalance, error) {
	var result BRC20TokenBalance
	err := db.RemoteMaster().Get(&result, "select * from balance where ticker = $1 and address = $2", ticker, address)
	if err != nil {
		if err == sql.ErrNoRows {
			return &BRC20TokenBalance{Ticker: ticker, Address: address}, nil
		}
		return nil, err
	}
	return &result, nil
}

func (b *BRC20TokenBalance) FindUserAllBalance(address string) ([]*BRC20TokenBalance, error) {
	var result []*BRC20TokenBalance
	err := db.Master().Select(&result, "select * from balance where address = $1", address)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}
	}
	return result, nil
}

func (b *BRC20TokenBalance) GetTickerHoldersMap(tickers []string) (map[string]int, error) {
	var result []struct {
		Ticker string `db:"ticker"`
		Num    int    `db:"num"`
	}
	res := make(map[string]int, 0)
	q, a, err := sqlx.In("select ticker, count(DISTINCT address) as num from balance where ticker in (?) group by ticker", tickers)
	if err != nil {
		return nil, err
	}
	q = db.Master().Rebind(q)
	err = db.Master().Select(&result, q, a...)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, nil
		}
		return nil, err
	}

	for _, v := range result {
		res[v.Ticker] = v.Num
	}

	return res, nil
}

func (b *BRC20TokenBalance) GetTickerHolders(ticker string) (int, error) {
	var result int
	err := db.Master().Get(&result, "select count(DISTINCT address) from balance where ticker = $1", ticker)
	if err != nil {
		return 0, err
	}
	return result, nil
}

type HolderEntity struct {
	Rank    int    `json:"rank"`
	Address string `json:"address"`
	Balance string `json:"balance"`
}

type HolderListResponse struct {
	PageResponse
	Data []HolderEntity `json:"data"`
}

func (b *BRC20TokenBalance) CountByTicker(ticker string) (int, error) {
	var result int
	err := db.Master().Get(&result, "select count(*) from balance where ticker = $1 and overall_balance > 0", ticker)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (b *BRC20TokenBalance) GetUserList(ticker string, pageNo int, pageCount int) (*HolderListResponse, error) {
	total, err := b.CountByTicker(ticker)
	if err != nil {
		return nil, err
	}
	var users []BRC20TokenBalance
	resp := &HolderListResponse{
		PageResponse: PageResponse{Total: total, CurrentPage: pageNo, PageSize: pageCount, TotalPages: (total + pageCount - 1) / pageCount},
	}
	err = db.Master().Select(&users, "select address, overall_balance from balance where ticker = $1 and overall_balance > 0 order by overall_balance DESC limit $2 offset $3", ticker, pageCount, (pageNo-1)*pageCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return resp, nil
		}
	}

	entities := make([]HolderEntity, 0)
	for i, v := range users {
		entities = append(entities, HolderEntity{
			Rank:    pageCount*(pageNo-1) + i + 1,
			Address: v.Address,
			Balance: v.OverallBalance,
		})
	}
	resp.Data = entities
	return resp, nil
}
