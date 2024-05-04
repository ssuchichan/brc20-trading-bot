package utils

import (
	"brc20-trading-bot/constant"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUTXO(t *testing.T) {
	pubKey, err := GetPubkeyFromAddress("fra1xe49fmdju4meynvamxe3arxn36vscpgfsrvjqzgx288m5cq6xc2q0ndjcy")
	assert.Nil(t, err)
	GetOwnedUTXO(pubKey, "https://prod-testnet.prod.findora.org:8668")
}

func TestSend(t *testing.T) {
	res, err := SendTx("burst sort child success muscle gaze salon swing orphan trim shaft climb",
		"Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=",
		"Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=",
		"2",
		"only",
		"2",
		constant.BRC20_OP_TRANSFER)
	assert.Nil(t, err)
	fmt.Println(res)
}

func TestSendRobotBatch(t *testing.T) {
	res, err := SendRobotBatch("burst sort child success muscle gaze salon swing orphan trim shaft climb")
	assert.Nil(t, err)
	fmt.Println(res)
}

func TestSendToAirDropAccount(t *testing.T) {
	pubkey, err := GetPubkeyFromAddress("fra1n4urmyxshvgy20arz99us2z6nlk32hherj6mpynwag2z56wunhlsevha2x")
	assert.Nil(t, err)
	res, err := Transfer("zoo nerve assault talk depend approve mercy surge bicycle ridge dismiss satoshi boring opera next fat cinnamon valley office actor above spray alcohol giant",
		pubkey,
		"20000000")
	assert.Nil(t, err)
	fmt.Println(res)
}
