package utils

import (
	"brc20-trading-bot/constant"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"math/big"
)

// GetAddressFromScript
func GetAddressFromScript(script string) (string, error) {
	result, err := base64.URLEncoding.DecodeString(script)
	if err != nil {
		return "", err
	}
	return bech32.EncodeFromBase256(constant.HRP, result)
}

func GetPubkeyFromAddress(addr string) (string, error) {
	hrp, data, err := bech32.DecodeToBase256(addr)
	if err != nil {
		return "", err
	}
	if hrp != constant.HRP {
		return "", errors.New("invalid address")
	}

	return base64.URLEncoding.EncodeToString(data), nil
}

func VerifyMessage(publicKey string, text string, sig string) bool {
	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return false
	}

	hash := magicHash(text)
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	return ecdsa.Verify(pubKey.ToECDSA(), hash[:], new(big.Int).SetBytes(sigBytes[1:33]), new(big.Int).SetBytes(sigBytes[33:]))
}

// varintBufNum returns the variable-length encoded representation of n.
func varintBufNum(n uint64) []byte {
	if n < 253 {
		return []byte{byte(n)}
	} else if n < 0x10000 {
		buf := make([]byte, 3)
		buf[0] = 253
		binary.LittleEndian.PutUint16(buf[1:], uint16(n))
		return buf
	} else if n < 0x100000000 {
		buf := make([]byte, 5)
		buf[0] = 254
		binary.LittleEndian.PutUint32(buf[1:], uint32(n))
		return buf
	}
	buf := make([]byte, 9)
	buf[0] = 255
	binary.LittleEndian.PutUint32(buf[1:], uint32(n&0xffffffff))
	binary.LittleEndian.PutUint32(buf[5:], uint32(n>>32))
	return buf
}

// magicHash calculates the magic hash of a message.
func magicHash(message string) []byte {
	const magicBytes = "Bitcoin Signed Message:\n"
	var magicBytesLength = len(magicBytes)

	prefix1 := varintBufNum(uint64(magicBytesLength))
	messageBytes := []byte(message)
	prefix2 := varintBufNum(uint64(len(messageBytes)))

	buf := make([]byte, len(prefix1)+len(magicBytes)+len(prefix2)+len(messageBytes))
	copy(buf, prefix1)
	copy(buf[len(prefix1):], magicBytes)
	copy(buf[len(prefix1)+len(magicBytes):], prefix2)
	copy(buf[len(prefix1)+len(magicBytes)+len(prefix2):], messageBytes)
	return hash256(buf)
}

// hash256 calculates the SHA-256 hash of data.
func hash256(data []byte) []byte {
	h := sha256.Sum256(data)
	h = sha256.Sum256(h[:])
	return h[:]
}
