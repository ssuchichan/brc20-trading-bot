package model

import (
	"brc20-trading-bot/db"
	"brc20-trading-bot/platform"
	"github.com/sirupsen/logrus"
	"time"
)

type Robot struct {
	Base
	Account    string `json:"account" db:"account"`
	PrivateKey string `json:"privateKey" db:"private_key"`
	Mnemonic   string `json:"mnemonic" db:"mnemonic"`
}

func (r *Robot) CreateBatch() (bool, error) {
	ok, err := r.IsCreated()
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	logrus.Info("Generating accounts...")
	var (
		robotList []*Robot
		robotsBuy []*Robot
	)
	for i := 0; i < 100; i++ {
		curMnemonic := platform.GetMnemonic()
		privateKey := platform.Mnemonic2PrivateKey([]byte(curMnemonic))
		curAccount := platform.Mnemonic2Bench32([]byte(curMnemonic))
		robotList = append(robotList, &Robot{
			Account:    curAccount,
			PrivateKey: privateKey,
			Mnemonic:   curMnemonic,
			Base: Base{
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
		})
	}
	for i := 0; i < 100; i++ {
		curMnemonic := platform.GetMnemonic()
		privateKey := platform.Mnemonic2PrivateKey([]byte(curMnemonic))
		curAccount := platform.Mnemonic2Bench32([]byte(curMnemonic))
		robotsBuy = append(robotsBuy, &Robot{
			Account:    curAccount,
			PrivateKey: privateKey,
			Mnemonic:   curMnemonic,
			Base: Base{
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
		})
	}

	_, err = db.Master().NamedExec("INSERT INTO robot_list (account, private_key, mnemonic, create_time, update_time) VALUES (:account, :private_key, :mnemonic, :create_time, :update_time)", robotList)
	_, err = db.Master().NamedExec("INSERT INTO robot_buy (account, private_key, mnemonic, create_time, update_time) VALUES (:account, :private_key, :mnemonic, :create_time, :update_time)", robotsBuy)

	return true, err
}

func (r *Robot) IsCreated() (bool, error) {
	var resultBuy, resultList int
	err := db.Master().Get(&resultBuy, "select count(*) from robot_buy")
	if err != nil {
		return false, err
	}
	err = db.Master().Get(&resultList, "select count(*) from robot_list")
	return resultBuy > 0 && resultList > 0, nil
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
	err := db.Master().Select(&result, "select account from robot where ty=0")
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Robot) AllBuyAccounts() ([]string, error) {
	var result []string
	err := db.Master().Select(&result, "select account from robot where ty=1")
	if err != nil {
		return nil, err
	}

	return result, nil
}
