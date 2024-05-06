package model

import (
	"brc20-trading-bot/constant"
	"brc20-trading-bot/db"
	"database/sql"

	"time"

	"github.com/jmoiron/sqlx"
)

type ListRecord struct {
	Base
	Ticker         string `json:"ticker,omitempty" db:"ticker"`
	User           string `json:"user,omitempty" db:"user"`
	Amount         string `json:"amount,omitempty" db:"amount"`
	Price          string `json:"price,omitempty" db:"price"`
	State          int    `json:"state,omitempty" db:"state"` // 0 挂单中 , 1 取消, 2 完成 3 待上架
	ToUser         string `json:"to_user" db:"to_user"`
	CenterMnemonic string `json:"center_mnemonic" db:"center_mnemonic"`
}

func (l *ListRecord) SumListAmount(addr string) (int64, error) {
	var total int64
	err := db.RemoteMaster().Get(&total, "SELECT sum(Amount) FROM list_record WHERE state=$1 AND user=$2", constant.ListFinished, addr)
	return total, err
}

func (l *ListRecord) InsertToDB() (int64, error) {
	if l.CreateTime == 0 {
		l.CreateTime = time.Now().Unix()
	}
	l.State = constant.ListWaiting
	l.UpdateTime = time.Now().Unix()

	var insertedID int64
	err := db.RemoteMaster().Get(&insertedID,
		"INSERT INTO list_record (ticker, \"user\", amount, price, state, create_time, update_time, center_mnemonic) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		l.Ticker, l.User, l.Amount, l.Price, l.State, l.CreateTime, l.UpdateTime, l.CenterMnemonic)
	if err != nil {
		return 0, err
	}

	return insertedID, nil
}

func (l *ListRecord) ConfirmList() error {
	l.UpdateTime = time.Now().Unix()
	l.State = constant.Listing
	_, err := db.RemoteMaster().NamedExec("update list_record set state = :state, update_time = :update_time where id = :id and \"user\" = :user", l)
	if err != nil {
		return err
	}
	return nil
}

func (l *ListRecord) Cancel() error {
	l.UpdateTime = time.Now().Unix()
	l.State = constant.ListCancel
	_, err := db.RemoteMaster().NamedExec("update list_record set state = :state, update_time = :update_time where id = :id and \"user\" = :user", l)
	if err != nil {
		return err
	}
	return nil
}

func (l *ListRecord) Finished() error {
	l.UpdateTime = time.Now().Unix()
	l.State = constant.ListFinished
	_, err := db.RemoteMaster().NamedExec("update list_record set state = :state, update_time = :update_time, to_user = :to_user where id = :id", l)
	if err != nil {
		return err
	}
	return nil
}

func (l *ListRecord) GetById(id int) (*ListRecord, error) {
	var result ListRecord
	err := db.Master().Get(&result, "select * from list_record where id = $1", id)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type MarketInfo struct {
	Ticker      string `db:"ticker"`
	FloorPrice  string `db:"floor_price"`
	TotalVal24h string `db:"total_val_24h"`
	TotalVal    string `db:"total_val"`
}

func (l *ListRecord) GetMarketInfo(id int) (*MarketInfo, error) {
	var result MarketInfo
	err := db.Master().Get(&result, `select 
	ticker, 
	min(price / amount) as floor_price, 
	SUM(CASE WHEN create_time >= EXTRACT(epoch FROM CURRENT_TIMESTAMP - INTERVAL '24 HOUR') THEN price ELSE 0 END) AS total_val_24h,
	sum(price) as total_val 
	from list_record a where state = 2 and id = $1 and amount != 0 group by ticker;`, id)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (l *ListRecord) GetMarketInfoMap(tickers []string) (map[string]*MarketInfo, error) {
	res := make(map[string]*MarketInfo)
	var infos []*MarketInfo
	q, a, err := sqlx.In(`select 
	ticker, 
	min(price / amount) as floor_price, 
	SUM(CASE WHEN create_time >= EXTRACT(epoch FROM CURRENT_TIMESTAMP - INTERVAL '24 HOUR') THEN price ELSE 0 END) AS total_val_24h,
	sum(price) as total_val 
	from list_record a where state = 2 and amount != 0 and ticker in (?) group by ticker;`, tickers)
	if err != nil {
		return nil, err
	}
	q = db.Master().Rebind(q)
	err = db.Master().Select(&infos, q, a...)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, nil
		}
		return nil, err
	}
	for _, v := range infos {
		res[v.Ticker] = v
	}

	return res, nil
}

type UserTickListRecord struct {
	Id         int    `json:"id"`
	Ticker     string `json:"ticker"`
	From       string `json:"from"`
	Amount     string `json:"amount"`
	Price      string `json:"price"`
	State      int    `json:"state"`
	To         string `json:"to"`
	CreateTime int64  `json:"create_time"`
}

type UserTickerListRecordsResponse struct {
	PageResponse
	Data []UserTickListRecord `json:"data"`
}

type UserTickerListFindParams struct {
	PageSearch
	User   string `db:"user"`
	Ticker string `db:"ticker"`
	State  int    `db:"state"`
}

func (l *ListRecord) CountByTickerAndUser(params UserTickerListFindParams) (int, error) {
	var result int
	sql := "select count(*) from list_record where true "
	if params.User != "" {
		sql += "And \"user\" = :user"
	}
	sql += " And ticker = :ticker and state = :state"
	rows, err := db.Master().NamedQuery(sql, params)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return 0, err
		}
	}

	return result, nil
}

func (l *ListRecord) FindPageList(pageNo int, pageCount int, params UserTickerListFindParams) (*UserTickerListRecordsResponse, error) {
	total, err := l.CountByTickerAndUser(params)
	if err != nil {
		return nil, err
	}
	res := &UserTickerListRecordsResponse{
		PageResponse: PageResponse{Total: total, CurrentPage: pageNo, PageSize: pageCount, TotalPages: (total + pageCount - 1) / pageCount},
	}
	limit, offset := pageCount, (pageNo-1)*pageCount
	var result []*ListRecord
	sql := "select * from list_record where true "
	if params.User != "" {
		sql += "And \"user\" = :user"
	}
	sql += " And ticker = :ticker and state = :state order by price ASC, id DESC"
	sql += " limit :limit offset :offset"
	params.Limit = limit
	params.Offset = offset
	rows, err := db.Master().NamedQuery(sql, params)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tmp ListRecord
		if err := rows.StructScan(&tmp); err != nil {
			return nil, err
		}
		result = append(result, &tmp)
	}

	var records []UserTickListRecord
	for _, v := range result {
		records = append(records, UserTickListRecord{
			Id:         int(v.Id),
			Ticker:     v.Ticker,
			Amount:     v.Amount,
			Price:      v.Price,
			From:       v.User,
			State:      v.State,
			CreateTime: v.CreateTime,
			To:         v.ToUser,
		})
	}
	res.Data = records
	return res, nil
}

func (l *ListRecord) CountOrderByTickerAndUser(params UserTickerListFindParams) (int, error) {
	var result int
	sql := "select count(*) from list_record where true "
	if params.User != "" {
		sql += "And \"user\" = :user"
	}
	sql += " And ticker = :ticker and (state = 0 or state = 2)"
	rows, err := db.Master().NamedQuery(sql, params)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return 0, err
		}
	}

	return result, nil
}

func (l *ListRecord) FindOrderPageList(pageNo int, pageCount int, params UserTickerListFindParams) (*UserTickerListRecordsResponse, error) {
	total, err := l.CountOrderByTickerAndUser(params)
	if err != nil {
		return nil, err
	}
	res := &UserTickerListRecordsResponse{
		PageResponse: PageResponse{Total: total, CurrentPage: pageNo, PageSize: pageCount, TotalPages: (total + pageCount - 1) / pageCount},
	}
	limit, offset := pageCount, (pageNo-1)*pageCount
	var result []*ListRecord
	sql := "select * from list_record where true "
	if params.User != "" {
		sql += "And \"user\" = :user"
	}
	sql += " And ticker = :ticker and (state = 0 or state = 2) order by price ASC, id DESC"
	sql += " limit :limit offset :offset"
	params.Limit = limit
	params.Offset = offset
	rows, err := db.Master().NamedQuery(sql, params)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tmp ListRecord
		if err := rows.StructScan(&tmp); err != nil {
			return nil, err
		}
		result = append(result, &tmp)
	}

	var records []UserTickListRecord
	for _, v := range result {
		records = append(records, UserTickListRecord{
			Id:         int(v.Id),
			Ticker:     v.Ticker,
			Amount:     v.Amount,
			Price:      v.Price,
			From:       v.User,
			State:      v.State,
			CreateTime: v.CreateTime,
			To:         v.ToUser,
		})
	}
	res.Data = records
	return res, nil
}

func (l *ListRecord) GetRobotListRecord() ([]*ListRecord, error) {
	r := &Robot{}
	robots, err := r.AllListAccounts()
	if err != nil {
		return nil, err
	}

	var result []*ListRecord
	q, a, err := sqlx.In("select * from list_record where state = 0 and \"user\" in (?) order by price DESC;", robots)
	if err != nil {
		return nil, err
	}

	q = db.RemoteMaster().Rebind(q)
	err = db.RemoteMaster().Select(&result, q, a...)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, nil
		}
		return nil, err
	}
	return result, nil
}
