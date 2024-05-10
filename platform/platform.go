package platform

/*
#cgo LDFLAGS: -L../target/release/ -lplatform -lm

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <stdlib.h>

const char *get_tx_str(
	char* remainPtr, uint32_t remainLen,
    char *fromSigPtr, uint32_t fromSigLen,
	char *receiverPtr, uint32_t receiverLen,
    char *toPtr, uint32_t toLen,
    char *transAmountPrt, uint32_t transAmountLen,
	char *urlPtr, uint32_t urlLen,
    char *tickPtr, uint8_t tickLen,
	char *fraPricePrt, uint32_t fraPriceLen,
	char *brcTypePtr, uint32_t brcTypeLen
);

uint64_t get_seq_id(char *urlPrt, uint32_t urlLen);

const char *generate_mnemonic_default();

const char *mnemonic_to_bench32(char *fromSigPtr, uint32_t fromSigLen);
const char* mnemonic_to_private_key(char* mnemonicPtr, uint32_t mnemonicLen);
const char* mnemonic_to_public_key(char* mnemonicPtr, uint32_t mnemonicLen);
const char* generate_private_key();

const char* private_key_to_bech32(char* skBase64Ptr, uint32_t keyLen);

const char *get_transfer_tx_str(
    char *fromSigPtr, uint32_t fromSigLen,
	char *receiverPtr, uint32_t receiverLen,
    char *fraPricePrt, uint32_t fraPriceLen,
	char *urlPtr, uint32_t urlLen
);

const char *get_send_robot_batch_tx(
	char *fromSigPtr, uint32_t fromSigLen,
	char *urlPtr, uint32_t urlLen
);

uint64_t get_user_fra_balance(
	char *fromSigPtr, uint32_t fromSigLen,
	char *urlPtr, uint32_t urlLen
);

*/
import "C"
import (
	"unsafe"
)

func GetTxBody(remain []byte, fromSig []byte, receiver []byte, to []byte, url []byte, transAmount []byte, tick []byte, fraPrice []byte, brcType []byte) string {
	// Call C function
	result := C.get_tx_str(
		(*C.char)(unsafe.Pointer(&remain[0])), C.uint32_t(len(remain)),
		(*C.char)(unsafe.Pointer(&fromSig[0])), C.uint32_t(len(fromSig)),
		(*C.char)(unsafe.Pointer(&receiver[0])), C.uint32_t(len(receiver)),
		(*C.char)(unsafe.Pointer(&to[0])), C.uint32_t(len(to)),
		(*C.char)(unsafe.Pointer(&transAmount[0])), C.uint32_t(len(transAmount)),
		(*C.char)(unsafe.Pointer(&url[0])), C.uint32_t(len(url)),
		(*C.char)(unsafe.Pointer(&tick[0])), C.uint8_t(len(tick)),
		(*C.char)(unsafe.Pointer(&fraPrice[0])), C.uint32_t(len(fraPrice)),
		(*C.char)(unsafe.Pointer(&brcType[0])), C.uint32_t(len(brcType)),
	)

	// Convert result to Go string
	resultStr := C.GoString(result)
	return resultStr
}

func GetTransferBody(fromSig []byte, receiver []byte, transAmount []byte, url []byte) string {
	// Call C function
	result := C.get_transfer_tx_str(
		(*C.char)(unsafe.Pointer(&fromSig[0])), C.uint32_t(len(fromSig)),
		(*C.char)(unsafe.Pointer(&receiver[0])), C.uint32_t(len(receiver)),
		(*C.char)(unsafe.Pointer(&transAmount[0])), C.uint32_t(len(transAmount)),
		(*C.char)(unsafe.Pointer(&url[0])), C.uint32_t(len(url)),
	)

	// Convert result to Go string
	resultStr := C.GoString(result)
	return resultStr
}

func GetSeqId(url []byte) uint64 {
	result := C.get_seq_id((*C.char)(unsafe.Pointer(&url[0])), C.uint32_t(len(url)))
	return uint64(result)
}

func GetMnemonic() string {
	result := C.generate_mnemonic_default()
	resultStr := C.GoString(result)
	return resultStr
}

func Mnemonic2Bench32(fromSig []byte) string {
	result := C.mnemonic_to_bench32(
		(*C.char)(unsafe.Pointer(&fromSig[0])), C.uint32_t(len(fromSig)))

	// Convert result to Go string
	resultStr := C.GoString(result)
	return resultStr
}

func Mnemonic2PrivateKey(mnemonic []byte) string {
	result := C.mnemonic_to_private_key((*C.char)(unsafe.Pointer(&mnemonic[0])), C.uint32_t(len(mnemonic)))
	str := C.GoString(result)
	return str
}

func Mnemonic2PublicKey(mnemonic []byte) string {
	result := C.mnemonic_to_public_key((*C.char)(unsafe.Pointer(&mnemonic[0])), C.uint32_t(len(mnemonic)))
	str := C.GoString(result)
	return str
}

func GeneratePrivateKey() string {
	result := C.generate_private_key()
	str := C.GoString(result)
	return str
}

func PrivateKey2Bech32(privateKey []byte) string {
	result := C.private_key_to_bech32((*C.char)(unsafe.Pointer(&privateKey[0])), C.uint32_t(len(privateKey)))
	str := C.GoString(result)
	return str
}

func GetSendRobotBatchTxBody(fromSig []byte, url []byte) string {
	result := C.get_send_robot_batch_tx(
		(*C.char)(unsafe.Pointer(&fromSig[0])), C.uint32_t(len(fromSig)),
		(*C.char)(unsafe.Pointer(&url[0])), C.uint32_t(len(url)),
	)

	// Convert result to Go string
	resultStr := C.GoString(result)
	return resultStr
}

func GetUserFraBalance(fromSig []byte, url []byte) uint64 {
	result := C.get_user_fra_balance(
		(*C.char)(unsafe.Pointer(&fromSig[0])), C.uint32_t(len(fromSig)),
		(*C.char)(unsafe.Pointer(&url[0])), C.uint32_t(len(url)),
	)

	return uint64(result)
}
