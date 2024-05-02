package platform

import (
	"brc20-trading-bot/constant"
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDemo(t *testing.T) {
	Demo()
}

func TestGetSeqId2(t *testing.T) {
	result := GetSeqId([]byte("https://prod-testnet.prod.findora.org:8668"))
	fmt.Println(result)
}

func TestGetTx(t *testing.T) {
	fromSig := []byte("burst sort child success muscle gaze salon swing orphan trim shaft climb")
	to := []byte("Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=")
	url := []byte("https://prod-testnet.prod.findora.org:8668")

	transAmount := []byte("2")
	tick := []byte("only")
	brcType := []byte(constant.BRC20_OP_TRANSFER)
	ans := GetTxBody(fromSig, to, to, url, transAmount, tick, []byte("2.2"), brcType)
	fmt.Println(ans)
}

func TestGetMintTx(t *testing.T) {
	fromSig := []byte("burst sort child success muscle gaze salon swing orphan trim shaft climb")
	to := []byte("Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=")
	url := []byte("https://prod-testnet.prod.findora.org:8668")

	transAmount := []byte("2")
	tick := []byte("only")
	brcType := []byte(constant.BRC20_OP_MINT)
	ans := GetTxBody(fromSig, to, to, url, transAmount, tick, []byte("2.2"), brcType)
	fmt.Println(ans)
}

func TestMnemonic(t *testing.T) {
	result := Mnemonic2Bench32([]byte("burst sort child success muscle gaze salon swing orphan trim shaft climb"))
	assert.Equal(t, "fra1ysa5953dcx6ldwufetpsc63hrt9xegm86735pwl9xpnlmt79r45srexf0r", result)
}

func TestGetUserFRABalance(t *testing.T) {
	result := GetUserFraBalance([]byte("thought faint misery file cube cage agent flight gallery bundle thrive grant whip pig then purchase movie essence obey old cup loud until goose"),
		[]byte("https://prod-testnet.prod.findora.org:8668"))
	fmt.Println(result)
}
