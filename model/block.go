package model

import (
	"brc20-trading-bot/db"
	"database/sql"

	"time"
)

type Block struct {
	Base
	Height     int64 `json:"height" db:"height"`
	IsFinished bool  `json:"is_finished" db:"is_finished"`
}

func NewBlockFromDB(h int64) (*Block, error) {
	var result Block
	err := db.Master().Get(&result, "select * from  block where height = $1", h)
	if err != nil {
		if err == sql.ErrNoRows {
			return &Block{Height: h}, nil
		}
		return nil, err
	}
	return &result, nil
}

func (b *Block) IsSyncFinished() bool {
	return b.IsFinished
}

func (b *Block) InsertToDB() error {
	if b.CreateTime == 0 {
		b.CreateTime = time.Now().Unix()
	}
	b.UpdateTime = time.Now().Unix()
	_, err := db.Master().NamedExec("INSERT INTO block (height, is_finished, create_time, update_time) values (:height, :is_finished, :create_time, :update_time)", b)
	if err != nil {
		return err
	}
	return nil
}

func (b *Block) FinishHandleBlock() error {
	b.UpdateTime = time.Now().Unix()
	b.IsFinished = true
	_, err := db.Master().NamedExec("update block set is_finished = :is_finished, update_time = :update_time where id = :id", b)
	if err != nil {
		return err
	}
	return nil
}

func (b *Block) LatestHeight() (int64, error) {
	var height int64
	err := db.Master().Get(&height, "select height from block order by id DESC limit 1")
	if err != nil {
		return 0, err
	}
	return height, nil
}

func GetLatestBlock() (*Block, error) {
	var block Block
	err := db.Master().Get(&block, "select * from block order by id DESC limit 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return &Block{}, nil
		}
		return nil, err
	}
	return &block, nil
}

func GetLatestFinishedBlock() (*Block, error) {
	var block Block
	err := db.Master().Get(&block, "select * from block where is_finished = 1 order by id DESC limit 1")
	if err != nil {
		if err == sql.ErrNoRows {
			return &Block{}, nil
		}
		return nil, err
	}
	return &block, nil
}
