package constant

// brc20 protocal
const (
	BRC20_P = "brc-20"
)

// brc20 op
const (
	BRC20_OP_DEPLOY   = "deploy"
	BRC20_OP_MINT     = "mint"
	BRC20_OP_TRANSFER = "transfer"
)

const (
	BRC20_OP_N_DEPLOY   = 0
	BRC20_OP_N_MINT     = 1
	BRC20_OP_N_TRANSFER = 2
)

// brc20 history
const (
	BRC20_HISTORY_TYPE_INSCRIBE_DEPLOY   = "inscribe-deploy"
	BRC20_HISTORY_TYPE_INSCRIBE_MINT     = "inscribe-mint"
	BRC20_HISTORY_TYPE_INSCRIBE_TRANSFER = "inscribe-transfer"
	BRC20_HISTORY_TYPE_TRANSFER          = "transfer"
	BRC20_HISTORY_TYPE_SEND              = "send"
	BRC20_HISTORY_TYPE_RECEIVE           = "receive"
)

const (
	BRC20_HISTORY_TYPE_N_INSCRIBE_DEPLOY   uint8 = 0
	BRC20_HISTORY_TYPE_N_INSCRIBE_MINT     uint8 = 1
	BRC20_HISTORY_TYPE_N_INSCRIBE_TRANSFER uint8 = 2
	BRC20_HISTORY_TYPE_N_TRANSFER          uint8 = 3
	BRC20_HISTORY_TYPE_N_SEND              uint8 = 4
	BRC20_HISTORY_TYPE_N_RECEIVE           uint8 = 5
)

var BRC20_HISTORY_TYPES_TO_N map[string]uint8 = map[string]uint8{
	BRC20_HISTORY_TYPE_INSCRIBE_DEPLOY:   BRC20_HISTORY_TYPE_N_INSCRIBE_DEPLOY,
	BRC20_HISTORY_TYPE_INSCRIBE_MINT:     BRC20_HISTORY_TYPE_N_INSCRIBE_MINT,
	BRC20_HISTORY_TYPE_INSCRIBE_TRANSFER: BRC20_HISTORY_TYPE_N_INSCRIBE_TRANSFER,
	BRC20_HISTORY_TYPE_TRANSFER:          BRC20_HISTORY_TYPE_N_TRANSFER,
	BRC20_HISTORY_TYPE_SEND:              BRC20_HISTORY_TYPE_N_SEND,
	BRC20_HISTORY_TYPE_RECEIVE:           BRC20_HISTORY_TYPE_N_RECEIVE,
}

var BRC20_HISTORY_TYPE_NAMES []string = []string{
	BRC20_HISTORY_TYPE_INSCRIBE_DEPLOY,
	BRC20_HISTORY_TYPE_INSCRIBE_MINT,
	BRC20_HISTORY_TYPE_INSCRIBE_TRANSFER,
	BRC20_HISTORY_TYPE_TRANSFER,
	BRC20_HISTORY_TYPE_SEND,
	BRC20_HISTORY_TYPE_RECEIVE,
}

var DEFAULT_DECIMAL_18 = "18"

// env
const (
	ENDPOINT      = "ENDPOINT"
	PlatInnerPort = "PLAT_INNER_PORT"
	PlatApiPort   = "PLAT_API_PORT"
	RedisUrl      = "REDISURL"
	RobotTick     = "ROBOT_TICK"
)

// hrp
const HRP = "fra"

// list status
// '0: 上架中 , 1: 取消, 2: 已完成, 3: 待上架';
const (
	Listing               = 0
	ListCancel            = 1
	ListFinished          = 2
	ListWaiting           = 3
	ListCancelToBeConfirm = 4
	ListToBeConfirm       = 5
	ListRefund            = 6
)

const TxMinFee = 10_000_000
