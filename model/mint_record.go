package model

import (
	"brc20-trading-bot/db"
	"database/sql"

	"time"

	"github.com/jmoiron/sqlx"
)

type MintRecord struct {
	Base
	Ticker string `json:"ticker,omitempty" db:"ticker"`
	User   string `json:"user,omitempty" db:"user"`
	Amount string `json:"amount,omitempty" db:"amount"`
}

func (m *MintRecord) MintTickerTotal(ticker string) (string, error) {
	var result struct {
		Total sql.NullString
	}
	err := db.Master().Get(&result, "select sum(amount) total from mint_record where ticker = $1", ticker)
	if err != nil {
		if err == sql.ErrNoRows {
			return "0", nil
		}
		return "0", err
	}

	if result.Total.String == "" {
		return "0", nil
	}

	return result.Total.String, nil
}

func (m *MintRecord) InsertToDB() error {
	if m.CreateTime == 0 {
		m.CreateTime = time.Now().Unix()
	}
	m.UpdateTime = time.Now().Unix()
	_, err := db.Master().NamedExec("INSERT INTO mint_record (ticker, \"user\", amount, create_time, update_time) values (:ticker, :user, :amount, :create_time, :update_time)", m)
	if err != nil {
		return err
	}
	return nil
}

func (m *MintRecord) GetTickerMintTotalMap(tickers []string) (map[string]string, error) {
	var result []struct {
		Ticker string `db:"ticker"`
		Num    string `db:"num"`
	}
	res := make(map[string]string, 0)
	q, a, err := sqlx.In("select ticker, SUM(amount) as num from mint_record where ticker in (?) group by ticker", tickers)
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
