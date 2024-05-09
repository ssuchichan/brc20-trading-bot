package utils

import (
	"brc20-trading-bot/constant"
	"brc20-trading-bot/model"
	"brc20-trading-bot/platform"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetOwnedUTXO(pubKey string, endpoint string) (sid uint64, record []byte, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/owned_utxos/%s", endpoint, pubKey))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	return
}

func SendTx(from string, receiverPubKey string, toPubKey string, amount string, tick string, fraPrice string, brcType string) (string, error) {
	txJsonString := platform.GetTxBody(
		[]byte(from),
		[]byte(receiverPubKey),
		[]byte(toPubKey),
		[]byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLAT_INNER_PORT))),
		[]byte(amount),
		[]byte(tick),
		[]byte(fraPrice),
		[]byte(brcType))
	if len(txJsonString) == 0 {
		return "", fmt.Errorf("insufficient FRA")
	}
	resultTx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(resultTx)
}

func sendRequest(resultTx string) (string, error) {
	req := model.NewBlockRequest(resultTx)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLAT_API_PORT)), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("request error")
	}

	return string(respBody), nil
}

func Transfer(from string, receiverPubKey string, amount string) (string, error) {
	txJsonString := platform.GetTransferBody([]byte(from), []byte(receiverPubKey), []byte(amount), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLAT_INNER_PORT))))
	if len(txJsonString) == 0 {
		return "", fmt.Errorf("insufficient FRA")
	}
	resultTx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(resultTx)
}

func SendRobotBatch(from string) (string, error) {
	txJsonString := platform.GetSendRobotBatchTxBody([]byte(from), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLAT_INNER_PORT))))
	resultTx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(resultTx)
}

func GetFraBalance(from string) uint64 {
	return platform.GetUserFraBalance([]byte(from), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PLAT_INNER_PORT))))
}
