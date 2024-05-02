package model

import (
	"brc20-trading-bot/db"
	"fmt"

	"time"
)

type AirdropRecord struct {
	Base
	FromUser string `json:"from_user,omitempty" db:"from_user"`
	ToUser   string `json:"to_user,omitempty" db:"to_user"`
	Amount   string `json:"amount,omitempty" db:"amount"`
}

func (a *AirdropRecord) InsertToDB() error {
	if a.CreateTime == 0 {
		a.CreateTime = time.Now().Unix()
	}
	a.UpdateTime = time.Now().Unix()
	isExist, err := a.IsExist(a.FromUser, a.ToUser)
	if err != nil {
		return err
	}
	if isExist {
		return fmt.Errorf("airdrop to %s exist", a.ToUser)
	}
	_, err = db.Master().NamedExec("INSERT INTO airdrop_record (from_user, to_user, amount, create_time, update_time) values (:from_user, :to_user, :amount, :create_time, :update_time)", a)
	if err != nil {
		return err
	}
	return nil
}

func (a *AirdropRecord) IsExist(from, to string) (bool, error) {
	var result int
	err := db.Master().Get(&result, "select count(*) from airdrop_record where from_user = $1 and to_user = $2", from, to)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}
