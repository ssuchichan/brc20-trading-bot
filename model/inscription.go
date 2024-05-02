package model

import (
	"brc20-trading-bot/constant"
	"encoding/json"
)

type InscriptionBRC20 struct {
	Proto        string `json:"p,omitempty"`
	Operation    string `json:"op,omitempty"`
	BRC20Tick    string `json:"tick,omitempty"`
	BRC20Max     string `json:"max,omitempty"`
	BRC20Amount  string `json:"amt,omitempty"`
	BRC20Limit   string `json:"lim,omitempty"` // option
	BRC20Decimal string `json:"dec,omitempty"` // option
}

func (body *InscriptionBRC20) Unmarshal(contentBody []byte) (err error) {
	var bodyMap map[string]interface{} = make(map[string]interface{}, 8)
	if err := json.Unmarshal(contentBody, &bodyMap); err != nil {
		return err
	}
	if v, ok := bodyMap["p"].(string); ok {
		body.Proto = v
	}
	if v, ok := bodyMap["op"].(string); ok {
		body.Operation = v
	}
	if v, ok := bodyMap["tick"].(string); ok {
		body.BRC20Tick = v
	}
	if v, ok := bodyMap["max"].(string); ok {
		body.BRC20Max = v
	}
	if v, ok := bodyMap["amt"].(string); ok {
		body.BRC20Amount = v
	}

	if _, ok := bodyMap["lim"]; !ok {
		body.BRC20Limit = body.BRC20Max
	} else {
		if v, ok := bodyMap["lim"].(string); ok {
			body.BRC20Limit = v
		}
	}

	if _, ok := bodyMap["dec"]; !ok {
		body.BRC20Decimal = constant.DEFAULT_DECIMAL_18
	} else {
		if v, ok := bodyMap["dec"].(string); ok {
			body.BRC20Decimal = v
		}
	}

	return nil
}
