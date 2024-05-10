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
	res, err := SendTx("", "burst sort child success muscle gaze salon swing orphan trim shaft climb",
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
	fmt.Println(pubkey)
	res, err := Transfer("7KPz-uvit2VO1WrdRMWzqZVorgIixz7fb3MR5QX3qcs=",
		pubkey,
		"2000000")
	assert.Nil(t, err)
	fmt.Println(res)
}

func TestAddressToPubKey(t *testing.T) {
	pk1, err := GetPubkeyFromAddress("fra1wnfcfahep40q2mjz6jw42ukylu72dp4c5rmcvvu4mwxn5urk825s8zfxpm") // dNOE9vkNXgVuQtSdVXLE_zymhrig94YzlduNOnB2Oqk=
	assert.Nil(t, err)
	fmt.Println(pk1)
	pk2, err := GetPubkeyFromAddress("fra1jfs43ad2m737dqp523s4v3j9sh4xmrum8s0uc4mah3qujpurkats8wcpq4") // kmFY9arfo-aANFRhVkZFheptj5s8H8xXfbxByQeDt1c=
	assert.Nil(t, err)
	fmt.Println(pk2)
	pk3, err := GetPubkeyFromAddress("fra1a8xal50pwjhl6z8u7e5rdjzgsj4fhhp8nsur8cc7t8p0fss6euks8msz52") // 6c3f0eF0r_0I_PZoNshIhKqb3CecODPjHlnC9MIazy0=
	assert.Nil(t, err)
	fmt.Println(pk3)
	pk4, err := GetPubkeyFromAddress("fra187p7pexx8z7kr5fza4tup48jkeldytj3yvsrwh8lw6a9fq45tzcs2x6nvw")
	assert.Nil(t, err)
	fmt.Println(pk4)
}
