package utils

import (
	"brc20-trading-bot/constant"
	_ "brc20-trading-bot/db"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAddressFromScript(t *testing.T) {
	account, err := GetAddressFromScript("cGaj6hH59eLXTgkepQb0sVxCEGSiPVyuOR-ziPbJX2Q=")
	assert.Nil(t, err)
	fmt.Println(account)
}

func Test_PublicKey(t *testing.T) {
	result, err := base64.StdEncoding.DecodeString("Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=")
	assert.Nil(t, err)
	answer, err := bech32.EncodeFromBase256(constant.HRP, result)
	assert.Nil(t, err)
	assert.Equal(t, "fra1xklsu8an2y4yd2e8kc43xspx54vrppp4l8fhtpey4n9z0kvmw2tqf76l2c", answer)
}

func TestGetPubkeyFromAddress(t *testing.T) {
	pubkey, err := GetPubkeyFromAddress("fra1xklsu8an2y4yd2e8kc43xspx54vrppp4l8fhtpey4n9z0kvmw2tqf76l2c")
	assert.Nil(t, err)
	assert.Equal(t, pubkey, "Nb8OH7NRKkarJ7YrE0AmpVgwhDX503WHJKzKJ9mbcpY=")
}

func TestUnisatSign(t *testing.T) {
	res := VerifyMessage("02e5ce539584735c77cdb53ce42a3468cfdb87f6c93cbd6b0fdfa790b03f338029", "hello world~", "H4WpsCzA/qKu+sTb72kZ+Smp9UdttkwzEC7dDbmmkuxCEuIconXu6OrJqHrr2Zc1EU/lqkWBUcUbZ7teqX+zp4Y=")
	fmt.Println(res)
}

func TestMagic(t *testing.T) {
	assert.Equal(t, "22d290bee19f60ac03256c95751334c8e5a3377394f702cef84910f9ff694503", hex.EncodeToString(magicHash("hello world~")))
}

func TestParse(t *testing.T) {
	sigBytes, _ := hex.DecodeString("304502210085a9b02cc0fea2aefac4dbef6919f929a9f5476db64c33102edd0db9a692ec42022012e21ca275eee8eac9a87aebd99735114fe5aa458151c51b67bb5ea97fb3a786")

	signature, err := ecdsa.ParseSignature(sigBytes)
	assert.Nil(t, err)
	_ = signature
}
