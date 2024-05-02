package model

import (
	"brc20-trading-bot/db"
	"brc20-trading-bot/decimal"
	"database/sql"
	"fmt"

	"time"
)

type Token struct {
	Ticker     string `json:"ticker,omitempty" db:"ticker"`
	Decimal    uint8  `json:"deciaml,omitempty" db:"dec"`
	Max        string `json:"max,omitempty" db:"max"`
	Limit      string `json:"limit,omitempty" db:"lim"`
	DeployUser string `json:"deploy_user,omitempty" db:"deploy_user"`
	Base
}

func NewTokenFromDBByTicker(ticker string) (*Token, error) {
	var result Token
	err := db.Master().Get(&result, "select * from token where ticker = $1", ticker)

	if err != nil {
		if err == sql.ErrNoRows {
			return &Token{Ticker: ticker}, nil
		}
		return nil, err
	}

	return &result, nil
}

func (t *Token) Exist(ticker string) (bool, error) {
	var result int
	err := db.Master().Get(&result, "select count(*) from token where ticker = $1", ticker)
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (t *Token) InsertToDB() error {
	if t.CreateTime == 0 {
		t.CreateTime = time.Now().Unix()
	}
	t.UpdateTime = time.Now().Unix()
	isExist, err := t.Exist(t.Ticker)
	if err != nil {
		return err
	}
	if isExist {
		return fmt.Errorf("%s exist", t.Ticker)
	}
	_, err = db.Master().NamedExec("INSERT INTO token (ticker, dec, max, lim, create_time, update_time, deploy_user) values (:ticker, :dec, :max, :lim, :create_time, :update_time, :deploy_user)", t)
	if err != nil {
		return err
	}
	return nil
}

func (t *Token) GetMax() (*decimal.Decimal, error) {
	result, _, err := decimal.NewDecimalFromString(t.Max)
	return result, err
}

func (t *Token) GetLimit() (*decimal.Decimal, error) {
	result, _, err := decimal.NewDecimalFromString(t.Limit)
	return result, err
}

func (t *Token) Count(typeFind int) (int, error) {
	switch typeFind {
	case 1:
		return t.countInProgress()
	case 2:
		return t.countComplete()
	default:
		return t.countAll()
	}
}

func (t *Token) countAll() (int, error) {
	var result int
	err := db.Master().Get(&result, "select count(*) from token")
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return result, nil
}

func (t *Token) countComplete() (int, error) {
	var result int
	err := db.Master().Get(&result, "select count(*) from token a left join (select ticker, sum(amount) amount from mint_record group by ticker) b on a.ticker = b.ticker where a.max = b.amount")
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return result, nil
}

func (t *Token) countInProgress() (int, error) {
	var result int
	err := db.Master().Get(&result, "select count(*) from token a left join (select ticker, sum(amount) amount from mint_record group by ticker) b on a.ticker = b.ticker or b.ticker is null where a.max != b.amount or b.amount is null")
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return result, nil
}

func (t *Token) GetById(id uint64) (*Token, error) {
	var result Token
	err := db.Master().Get(&result, "select * from token where id = $1", id)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type TokenListResponse struct {
	PageResponse
	Data []TokenList `json:"data"`
}

type TokenList struct {
	Id         uint64 `json:"id"`
	Ticker     string `json:"ticker"`
	DeployTime int64  `json:"deploy_time"`
	MintTotal  string `json:"mint_total"`
	Limit      string `json:"limit"`
	Decimal    uint8  `json:"decimal"`
	Max        string `json:"max"`
	Holders    int    `json:"holders"`
}

type FindParams struct {
	Type   int    `json:"type"` // 0 all, 1 inprogress, 2 complete
	Ticker string `json:"ticker"`
}

func (t *Token) findAllPageList(limit int, offset int, ticker string) ([]Token, error) {
	var (
		tokens []Token
		err    error
	)
	if ticker != "" {
		err = db.Master().Select(&tokens, "select * from token where ticker = $1 order by id desc limit $2 offset $3", ticker, limit, offset)
	} else {
		err = db.Master().Select(&tokens, "select * from token order by id desc limit $1 offset $2", limit, offset)
	}

	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (t *Token) findCompletePageList(limit, offset int, ticker string) ([]Token, error) {
	var (
		tokens []Token
		err    error
	)
	if ticker != "" {
		err = db.Master().Select(&tokens, "select a.* from token a left join (select ticker, sum(amount) amount from mint_record group by ticker) b on a.ticker = b.ticker where a.max = b.amount and a.ticker = $1 order by a.id DESC limit $2 offset $3", ticker, limit, offset)
	} else {
		err = db.Master().Select(&tokens, "select a.* from token a left join (select ticker, sum(amount) amount from mint_record group by ticker) b on a.ticker = b.ticker where a.max = b.amount order by a.id DESC limit $1 offset $2", limit, offset)
	}

	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (t *Token) findInprogressPageList(limit, offset int, ticker string) ([]Token, error) {
	var (
		tokens []Token
		err    error
	)
	if ticker != "" {
		err = db.Master().Select(&tokens, `select a.* from token a left join (select ticker, sum(amount) amount from mint_record group by ticker) b
		on a.ticker = b.ticker or b.ticker is null where (a.max != b.amount or b.amount is null) and a.ticker = $1 order by a.id DESC limit $2 offset $3;`, ticker, limit, offset)
	} else {
		err = db.Master().Select(&tokens, `select a.* from token a left join (select ticker, sum(amount) amount from mint_record group by ticker) b
		on a.ticker = b.ticker or b.ticker is null where a.max != b.amount or b.amount is null order by a.id DESC limit $1 offset $2;`, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (t *Token) FindPageList(pageNo int, pageCount int, params FindParams) (*TokenListResponse, error) {
	total, err := t.Count(params.Type)
	if err != nil {
		return nil, err
	}
	res := &TokenListResponse{
		PageResponse: PageResponse{Total: total, CurrentPage: pageNo, PageSize: pageCount, TotalPages: (total + pageCount - 1) / pageCount},
	}
	limit, offset := pageCount, (pageNo-1)*pageCount
	var tokens []Token
	switch params.Type {
	case 1:
		tokens, err = t.findInprogressPageList(limit, offset, params.Ticker)
	case 2:
		tokens, err = t.findCompletePageList(limit, offset, params.Ticker)
	default:
		tokens, err = t.findAllPageList(limit, offset, params.Ticker)
	}
	if err != nil {
		return nil, err
	}
	tokenTickers := make([]string, 0)
	for _, v := range tokens {
		tokenTickers = append(tokenTickers, v.Ticker)
	}
	if len(tokenTickers) == 0 {
		return res, nil
	}

	b := &BRC20TokenBalance{}
	holdersMap, err := b.GetTickerHoldersMap(tokenTickers)
	if err != nil {
		return nil, err
	}

	m := &MintRecord{}
	mintMap, err := m.GetTickerMintTotalMap(tokenTickers)
	if err != nil {
		return nil, err
	}

	var tokenList []TokenList
	for _, v := range tokens {
		mintTotal := mintMap[v.Ticker]
		if mintTotal == "" {
			mintTotal = "0"
		}
		tokenList = append(tokenList, TokenList{
			Id:         v.Id,
			Ticker:     v.Ticker,
			DeployTime: v.CreateTime,
			Max:        v.Max,
			MintTotal:  mintTotal,
			Holders:    holdersMap[v.Ticker],
			Limit:      v.Limit,
			Decimal:    v.Decimal,
		})
	}
	res.Data = tokenList
	return res, nil
}

type TokenDetailResponse struct {
	Ticker     string `json:"ticker"`
	DeployTime int64  `json:"deploy_time"`
	Max        string `json:"max"`
	Limit      string `json:"limit"`
	Decimal    uint8  `json:"decimal"`
	DeployUser string `json:"deploy_user"`
	Holders    int    `json:"holders"`
	MintTotal  string `json:"mint_total"`
}

func (t *Token) GetDetail(id uint64) (*TokenDetailResponse, error) {

	token, err := t.GetById(id)
	if err != nil {
		return nil, err
	}
	m := &MintRecord{}
	mintTotal, err := m.MintTickerTotal(token.Ticker)
	if err != nil {
		return nil, err
	}
	b := &BRC20TokenBalance{}
	hodlers, err := b.GetTickerHolders(token.Ticker)
	if err != nil {
		return nil, err
	}

	resp := &TokenDetailResponse{
		Ticker:     token.Ticker,
		DeployTime: token.CreateTime,
		Max:        token.Max,
		Limit:      token.Limit,
		Decimal:    token.Decimal,
		MintTotal:  mintTotal,
		Holders:    hodlers,
		DeployUser: token.DeployUser,
	}
	return resp, nil
}

type TokenCheckResponse struct {
	Ticker     string `json:"ticker"`
	DeployTime int64  `json:"deploy_time"`
	Max        string `json:"max"`
	Limit      string `json:"limit"`
	Decimal    uint8  `json:"decimal"`
	MintTotal  string `json:"mint_total"`
	IsExist    bool   `json:"is_exist"`
}

func (t *Token) CheckTicker(ticker string) (*TokenCheckResponse, error) {

	token, err := NewTokenFromDBByTicker(ticker)
	if err != nil {
		return nil, err
	}
	m := &MintRecord{}
	mintTotal, err := m.MintTickerTotal(token.Ticker)
	if err != nil {
		return nil, err
	}

	resp := &TokenCheckResponse{
		Ticker:     token.Ticker,
		DeployTime: token.CreateTime,
		Max:        token.Max,
		Limit:      token.Limit,
		Decimal:    token.Decimal,
		MintTotal:  mintTotal,
		IsExist:    !token.IsEmpty(),
	}
	return resp, nil
}

type TokenMarketListResponse struct {
	PageResponse
	Data []TokenMarketInfo `json:"data"`
}

type TokenMarketInfo struct {
	Ticker      string `json:"ticker"`
	Holders     int    `json:"holders"`
	FloorPrice  string `json:"floor_price"`
	TotalVal24h string `json:"total_val_24h"`
	TotalVal    string `json:"total_val"`
}

type MarketSearchParam struct {
	PageSearch
	Ticker string `db:"ticker"`
}

func (l *Token) countMarket(params MarketSearchParam) (int, error) {
	var result int
	sql := "select count(distinct a.ticker) from token a left join list_record b on a.ticker = b.ticker  where true "
	if params.Ticker != "" {
		sql += "And a.ticker like :ticker"
	}
	sql += " and b.state = 0"

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

func (t *Token) FindMarketInfos(pageNo int, pageCount int, params MarketSearchParam) (*TokenMarketListResponse, error) {
	total, err := t.countMarket(params)
	if err != nil {
		return nil, err
	}
	res := &TokenMarketListResponse{
		PageResponse: PageResponse{Total: total, CurrentPage: pageNo, PageSize: pageCount, TotalPages: (total + pageCount - 1) / pageCount},
	}
	limit, offset := pageCount, (pageNo-1)*pageCount
	var tickers []string
	sql := "select a.ticker from token a left join list_record b on a.ticker = b.ticker  where true "
	if params.Ticker != "" && params.Ticker != "%" {
		sql += "And a.ticker like :ticker"
	}
	sql += " and b.state = 0 group by a.ticker order by MAX(a.id) DESC"
	sql += " limit :limit offset :offset"
	params.Limit = limit
	params.Offset = offset
	rows, err := db.Master().NamedQuery(sql, params)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tmp string
		if err := rows.Scan(&tmp); err != nil {
			return nil, err
		}
		tickers = append(tickers, tmp)
	}

	if len(tickers) == 0 {
		return res, nil
	}

	l := &ListRecord{}
	infoMap, err := l.GetMarketInfoMap(tickers)
	if err != nil {
		return nil, err
	}

	b := &BRC20TokenBalance{}
	holdersMap, err := b.GetTickerHoldersMap(tickers)
	if err != nil {
		return nil, err
	}

	var data []TokenMarketInfo
	for _, ticker := range tickers {
		totalVal := "0"
		totalVal24 := "0"
		floorPrice := "0"
		if v, ok := infoMap[ticker]; ok {
			totalVal = v.TotalVal
			totalVal24 = v.TotalVal24h
			floorPrice = v.FloorPrice
		}

		data = append(data, TokenMarketInfo{
			Ticker:      ticker,
			TotalVal24h: totalVal24,
			TotalVal:    totalVal,
			FloorPrice:  floorPrice,
			Holders:     holdersMap[ticker],
		})
	}
	res.Data = data

	return res, nil
}
