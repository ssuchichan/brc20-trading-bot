package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertListRecord(t *testing.T) {
	record := &ListRecord{
		Ticker:         "test",
		User:           "test",
		Amount:         "1000",
		Price:          "2.34",
		State:          0,
		CenterMnemonic: "",
	}
	insertId, err := record.InsertToDB()
	assert.Nil(t, err)
	fmt.Println(insertId)
}

func TestGetById(t *testing.T) {
	r := &ListRecord{}
	record, err := r.GetById(1)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), record.Id)
}

func TestCancel(t *testing.T) {
	r := &ListRecord{}
	record, err := r.GetById(1)
	assert.Nil(t, err)
	err = record.Cancel()
	assert.Nil(t, err)
	record, err = r.GetById(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, record.State)
}

func TestFinished(t *testing.T) {
	r := &ListRecord{}
	record, err := r.GetById(1)
	assert.Nil(t, err)
	err = record.Finished()
	assert.Nil(t, err)
	record, err = r.GetById(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, record.State)
}

func TestFind(t *testing.T) {
	r := &ListRecord{}
	result, err := r.FindPageList(1, 10, UserTickerListFindParams{
		Ticker: "test",
		State:  1,
	})
	assert.Nil(t, err)
	fmt.Println(result)
}

func TestFindMarketInfos(t *testing.T) {
	r := &ListRecord{}
	result, err := r.GetMarketInfoMap([]string{"test", "test2"})
	assert.Nil(t, err)
	for k, v := range result {
		fmt.Println(k, v)
	}
}

func TestGetRobotListRecord(t *testing.T) {
	r := &ListRecord{}
	_, err := r.GetRobotListRecord()
	assert.Nil(t, err)

}
