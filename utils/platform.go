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
	"strings"
)

func GetOwnedUTXO(pubKey string, endpoint string) (sid uint64, record []byte, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/owned_utxos/%s", endpoint, pubKey))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	return
}

func SendTx(remain string, from string, receiverPubKey string, toPubKey string, amount string, tick string, fraPrice string, brcType string) (string, error) {
	txJsonString := platform.GetTxBody(
		[]byte(remain),
		[]byte(from),
		[]byte(receiverPubKey),
		[]byte(toPubKey),
		[]byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PlatInnerPort))),
		[]byte(amount),
		[]byte(tick),
		[]byte(fraPrice),
		[]byte(brcType))
	if strings.Contains(txJsonString, "error") || strings.Contains(txJsonString, "insufficient") {
		return "", errors.New(txJsonString)
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
	reqURL := fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PlatApiPort))
	resp, err := http.Post(reqURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request error: %v, %v", resp.Status, reqURL)
	}

	return string(respBody), nil
}

func Transfer(from string, receiverPubKey string, amount string) (string, error) {
	txJsonString := platform.GetTransferBody([]byte(from), []byte(receiverPubKey), []byte(amount), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PlatInnerPort))))
	if len(txJsonString) == 0 {
		return "", fmt.Errorf("insufficient FRA")
	}
	resultTx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(resultTx)
}

func SendRobotBatch(from string) (string, error) {
	txJsonString := platform.GetSendRobotBatchTxBody([]byte(from), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PlatInnerPort))))
	resultTx := base64.URLEncoding.EncodeToString([]byte(txJsonString))
	return sendRequest(resultTx)
}

func GetFraBalance(from string) uint64 {
	return platform.GetUserFraBalance([]byte(from), []byte(fmt.Sprintf("%s:%s", os.Getenv(constant.ENDPOINT), os.Getenv(constant.PlatInnerPort))))
}
