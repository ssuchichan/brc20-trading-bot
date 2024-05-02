package model

import (
	"encoding/base64"
	"encoding/json"
)

type Amount struct {
	NonConfidential string `json:"NonConfidential"`
}

type Input struct {
	Amount    Amount `json:"amount"`
	PublicKey string `json:"public_key"`
}

type Output struct {
	Amount    Amount            `json:"amount"`
	PublicKey string            `json:"public_key"`
	Memo      *InscriptionBRC20 `json:"memo,omitempty"`
}

type Transfer struct {
	Inputs []Input `json:"inputs"`
}

type BodyOutputRecord struct {
	PublicKey string `json:"public_key"`
	Amount    Amount `json:"amount"`
}

type BodyOutput struct {
	Record BodyOutputRecord `json:"record"`
	Memo   string           `json:"memo,omitempty"`
}

type TransferBody struct {
	Outputs  []BodyOutput `json:"outputs"`
	Transfer Transfer     `json:"transfer"`
}

type TransferAsset struct {
	Body TransferBody `json:"body"`
}

type Operation struct {
	TransferAsset TransferAsset `json:"TransferAsset"`
}

type Body struct {
	Operations []Operation `json:"operations"`
}

type HanldeBody struct {
	Body Body `json:"body"`
}

type BlockHeader struct {
	Height string `json:"height"`
}

type BlockData struct {
	Txs []string `json:"txs"`
}

type BlockInfo struct {
	Header BlockHeader `json:"header"`
	Data   BlockData   `json:"data"`
}

type BlockResult struct {
	Block BlockInfo `json:"block"`
}

type BlockResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	Result  *BlockResult   `json:"result,omitempty"`
	Error   *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func NewHanldeBodyFromResponse(resp string) (*HanldeBody, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(resp)
	if err != nil {
		return nil, err
	}
	var result HanldeBody
	err = json.Unmarshal(decodeBytes, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type PageResponse struct {
	Total       int `json:"total"`
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalPages  int `json:"totalPages"`
}

type PageSearch struct {
	Limit  int `db:"limit"`
	Offset int `db:"offset"`
}

type RequestParam struct {
	Tx string `json:"tx"`
}

type BlockRequest struct {
	Jsonrpc string       `json:"jsonrpc"`
	Id      string       `json:"id"`
	Method  string       `json:"method"`
	Params  RequestParam `json:"params"`
}

func NewBlockRequest(tx string) *BlockRequest {
	return &BlockRequest{
		Jsonrpc: "2.0",
		Id:      "anything",
		Method:  "broadcast_tx_sync",
		Params:  RequestParam{Tx: tx},
	}
}

type TxResult struct {
	Hash   string `json:"hash"`
	Height string `json:"height"`
	Tx     string `json:"tx"`
}

type TxResponse struct {
	Jsonrpc string   `json:"jsonrpc"`
	Result  TxResult `json:"result"`
}

type UnisatBalance struct {
	Address                   string `json:"address"`
	Satoshi                   int    `json:"satoshi"`
	PendingSatoshi            int    `json:"pendingSatoshi"`
	UtxoCount                 int    `json:"utxoCount"`
	BtcSatoshi                int    `json:"btcSatoshi"`
	BtcPendingSatoshi         int    `json:"btcPendingSatoshi"`
	BtcUtxoCount              int    `json:"btcUtxoCount"`
	InscriptionSatoshi        int    `json:"inscriptionSatoshi"`
	InscriptionPendingSatoshi int    `json:"inscriptionPendingSatoshi"`
	InscriptionUtxoCount      int    `json:"inscriptionUtxoCount"`
}

type UnisatTickerInfo struct {
	Ticker         string `json:"ticker"`
	OverallBalance string `json:"overallBalance"`
}

type UnisatTickerSummaryDetail struct {
	Ticker         string `json:"ticker"`
	OverallBalance string `json:"overallBalance"`
}

type UnisatTickerSummary struct {
	Height int                         `json:"height"`
	Start  int                         `json:"start"`
	Total  int                         `json:"total"`
	Detail []UnisatTickerSummaryDetail `json:"detail"`
}

type UnisatData interface {
	UnisatBalance | UnisatTickerInfo | UnisatTickerSummary
}

type UnisatResponse[T UnisatData] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
