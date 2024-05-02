package model

import (
	"brc20-trading-bot/db"
	"brc20-trading-bot/platform"
	"time"
)

type Robot struct {
	Base
	Mnemonic string `json:"mnemonic" db:"mnemonic"`
	Account  string `json:"account" db:"account"`
}

func (r *Robot) CreateBatch() (bool, error) {
	ok, err := r.IsCreated()
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	var robots []*Robot
	for i := 0; i < 200; i++ {
		curMnemonic := platform.GetMnemonic()
		curAccount := platform.Mnemonic2Bench32([]byte(curMnemonic))
		robots = append(robots, &Robot{
			Mnemonic: curMnemonic,
			Account:  curAccount,
			Base: Base{
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
		})
	}
	_, err = db.Master().NamedExec("INSERT INTO robot (account, mnemonic, create_time, update_time) VALUES (:account, :mnemonic, :create_time, :update_time)", robots)
	return true, err
}

func (r *Robot) IsCreated() (bool, error) {
	var result int
	err := db.Master().Get(&result, "select count(*) from robot")
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (r *Robot) GetById(id uint64) (*Robot, error) {
	var res Robot
	err := db.Master().Get(&res, "select * from robot where id = $1", id)
	return &res, err
}

func (r *Robot) Next() (*Robot, error) {
	var res Robot
	nextID := (r.Id + 2) % 200
	err := db.Master().Get(&res, "select * from robot where id = $1", nextID)
	return &res, err
}

func (r *Robot) AllListAcounts() ([]string, error) {
	var result []string
	err := db.Master().Select(&result, "select account from robot where id%2=1")
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Robot) AllBuyAccounts() ([]string, error) {
	var result []string
	err := db.Master().Select(&result, "select account from robot where id%2=0")
	if err != nil {
		return nil, err
	}

	return result, nil
}
