package main

import (
	"brc20-trading-bot/constant"
	"brc20-trading-bot/db"
	"brc20-trading-bot/decimal"
	"brc20-trading-bot/model"
	"brc20-trading-bot/platform"
	"brc20-trading-bot/utils"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type TendermintResult struct {
	Code      int    `json:"code"`
	Data      string `json:"data"`
	Log       string `json:"log"`
	Codespace string `json:"codespace"`
	Hash      string `json:"hash"`
}

type RpcResult struct {
	Jsonrpc string           `json:"jsonrpc"`
	ID      string           `json:"id"`
	Result  TendermintResult `json:"result"`
}

func initRobot() error {
	logrus.Info("Init accounts...")
	// 不创建账号，提前导入到数据库
	//r := &model.Robot{}
	//_, err := r.CreateBatch()
	//if err != nil {
	//	logrus.Info("Init robot accounts error: ", err)
	//	return err
	//}
	//if ok {
	//	// 第一次创建robot, 需要转钱
	//	// 通过发空投的账户给机器人打钱
	//	logrus.Info("Sending FRA...")
	//	_, err = utils.SendRobotBatch(os.Getenv(constant.AIRDROP_MNEMONIC))
	//	return err
	//}
	logrus.Info("Init accounts...ok")
	return nil
}

func init() {
	var (
		firstRobotListID uint64
		firstRobotBuyID  uint64
		err              error
	)

	if err = initRobot(); err != nil {
		logrus.Error(err)
		panic(err)
	}

	r := &model.Robot{}
	firstRobotListID, err = r.GetFirstRobotList()
	if err != nil {
		logrus.Fatal("Get first listRobot ID: ", err)
	}
	firstRobotBuyID, err = r.GetFirstRobotBuy()
	if err != nil {
		logrus.Fatal("Get first buyRobot ID: ", err)
	}
	logrus.Infof("The first listRobotId: %d, the first buyRobotId: %d", firstRobotListID, firstRobotBuyID)

	// S:L:M:R means Latest Mint Robot
	//db.MRedis().SetNX(context.Background(), "S:L:M:R", 1, time.Duration(0))
	// Least List Robot
	db.MRedis().SetNX(context.Background(), "S:L:L:R", firstRobotListID, time.Duration(0))
	// Least Buy Robot
	db.MRedis().SetNX(context.Background(), "S:L:B:R", firstRobotBuyID, time.Duration(0))

}

func main() {
	var (
		floorPrices         []string
		priceIndex          int64
		listInterval        int64
		buyInterval         int64
		priceUpdateInterval int64
		listLimit           int64
		listAmount          int64
		err                 error
	)
	floorPricesStr := os.Getenv("FLOOR_PRICES")
	prices := strings.Split(floorPricesStr, ",")
	for i := 0; i < len(prices); i++ {
		floorPrices = append(floorPrices, prices[i])
	}
	token := os.Getenv("ROBOT_TICK")
	listLimit, err = strconv.ParseInt(os.Getenv("LIST_LIMIT"), 10, 64)
	priceIndex, err = strconv.ParseInt(os.Getenv("PRICE_START_INDEX"), 10, 64)
	listInterval, err = strconv.ParseInt(os.Getenv("LIST_INTERVAL"), 10, 64)
	buyInterval, err = strconv.ParseInt(os.Getenv("BUY_INTERVAL"), 10, 64)
	priceUpdateInterval, err = strconv.ParseInt(os.Getenv("FLOOR_PRICE_UPDATE_INTERVAL"), 10, 64)
	listAmount, err = strconv.ParseInt(os.Getenv("LIST_AMOUNT"), 10, 64)
	if err != nil {
		logrus.Fatal(err)
	}
	d := int(priceUpdateInterval / 5) // *0.2
	r := rand.Intn(2*d+1) - d
	priceUpdateInterval = priceUpdateInterval + int64(r)

	logrus.Info("Tick: ", token)
	logrus.Info("Floor prices: ", floorPrices)
	logrus.Info("Current floor price: ", floorPrices[priceIndex])
	logrus.Infof("Floor prices updating interval: %ds", priceUpdateInterval)
	logrus.Infof("List limit: %v, list amount: %v", listLimit, listAmount)
	logrus.Infof("List interval: %ds, buy interval: %ds", listInterval, buyInterval)

	//mintTicker := time.NewTicker(60 * time.Second)
	addListTicker := time.NewTicker(time.Duration(listInterval) * time.Second)
	buyTicker := time.NewTicker(time.Duration(buyInterval) * time.Second)
	priceTicker := time.NewTicker(time.Duration(priceUpdateInterval) * time.Second)
	defer func() {
		//mintTicker.Stop()
		addListTicker.Stop()
		buyTicker.Stop()
		priceTicker.Stop()
	}()

	for {
		select {
		//case <-mintTicker.C:
		//	err := mint()
		//	if err != nil {
		//		utils.GetLogger().Errorf("mint tick err:%v", err)
		//		continue
		//	}
		case <-addListTicker.C:
			r := &model.Robot{}
			firstRobotID, err := r.GetFirstRobotList()
			if err != nil {
				logrus.Error("Get the first listRobotId: ", err)
				continue
			}
			robotCount, _ := r.GetListRobotCount()
			if robotCount == 0 {
				logrus.Error("The listRobot count is 0")
				continue
			}
			curFloorPrice := floorPrices[priceIndex]
			err = addList(curFloorPrice, listLimit, listAmount, int64(firstRobotID), int64(robotCount), token)
			if err != nil {
				utils.GetLogger().Errorf("list tick err: %v", err)
				continue
			}
		case <-buyTicker.C:
			r := &model.Robot{}
			firstRobotID, err := r.GetFirstRobotBuy()
			if err != nil {
				logrus.Error("Get the first listRobotId: ", err)
				continue
			}
			robotCount, _ := r.GetBuyRobotCount()
			if robotCount == 0 {
				logrus.Error("The buyRobot count is 0")
				continue
			}

			curFloorPrice := floorPrices[priceIndex]
			err = buy(curFloorPrice, int64(firstRobotID), int64(robotCount), token)
			if err != nil {
				utils.GetLogger().Errorf("buy tick err: %v", err)
				continue
			}
		case <-priceTicker.C:
			if priceIndex+1 == int64(len(prices)) {
				logrus.Info("Reached the last floor price, exit.")
				return
			}
			priceIndex += 1
			logrus.Info("Update floor price to: ", floorPrices[priceIndex])
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func mint() error {
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:M:R").Int64()
	if err != nil {
		return err
	}
	logrus.Infof("mint %d", latestRobotId)
	defer func() {
		db.MRedis().Set(context.Background(), "S:L:M:R", latestRobotId%200+1, time.Duration(0))
	}()

	r := &model.Robot{}
	curRobot, err := r.GetById(uint64(latestRobotId))
	if err != nil {
		fmt.Println(1)
		return err
	}

	curPubkey, err := utils.GetPubkeyFromAddress(curRobot.Account)
	if err != nil {
		return err
	}

	tick := os.Getenv(constant.RobotTick)

	token, err := model.NewTokenFromDBByTicker(tick)
	if err != nil {
		fmt.Println(2)
		return err
	}

	if token.IsEmpty() {
		return fmt.Errorf("%s not deployed", tick)
	}

	_, err = isMintFinished(token)
	if err != nil {
		return err
	}

	_, err = utils.SendTx("0", curRobot.Mnemonic, curPubkey, curPubkey, token.Limit, os.Getenv(constant.RobotTick), "0", constant.BRC20_OP_MINT)
	return err
}

func isMintFinished(token *model.Token) (bool, error) {
	m := &model.MintRecord{}
	total, err := m.MintTickerTotal(token.Ticker)
	if err != nil {
		return false, err
	}

	totalDecimal, _, err := decimal.NewDecimalFromString(total)
	if err != nil {
		return false, err
	}

	maxDecimal, _, err := decimal.NewDecimalFromString(token.Max)
	if err != nil {
		return false, err
	}

	if maxDecimal.Cmp(totalDecimal) <= 0 {
		return false, fmt.Errorf("%s have mint 100%%", token.Ticker)
	}
	return true, nil
}

func addList(floorPrice string, listLimit int64, listAmount int64, firstRobotID int64, robotCount int64, ticker string) error {
	// parse floor price
	decFloorPrice, _, err := decimal.NewDecimalFromString(floorPrice)
	if err != nil {
		return err
	}

	// S:L:L:R means Latest List Robot
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:L:R").Int64()
	if err != nil {
		logrus.Error("[List] redis: ", err)
		return err
	}
	logrus.Info("[List] the listRobot id: ", latestRobotId)
	// 1. 获取当前robot
	rb := &model.Robot{}
	curRobot, err := rb.GetRobotListById(uint64(latestRobotId))
	if err != nil {
		logrus.Error("[List] get listRobot id: ", err)
		return err
	}

	defer func() {
		db.MRedis().Set(context.Background(), "S:L:L:R", (latestRobotId+1)%robotCount+firstRobotID, time.Duration(0))
	}()

	// 当前list总量
	rec := &model.ListRecord{}
	totalList, err := rec.SumListAmount(ticker, 1_000_000*1.2*decFloorPrice.Float64())
	if err != nil {
		logrus.Error("[List] get list sum: ", err)
		return err
	}
	if totalList >= listLimit {
		logrus.Infof("[List] reach list limit, total listed: %d, list limit: %d", totalList, listLimit)
		return nil
	}
	logrus.Info("[List] current total list: ", totalList)

	// 单价上下浮动20%
	r := rand.Intn(41) - 20 // [-20, 20]随机数
	rate := float64(100+r) / 100.0
	unitPrice := decFloorPrice.Float64() * rate // 实际单价
	logrus.Infof("[List] expect floorPirce: %v, real unitPrice: %v, rate: %v", floorPrice, unitPrice, rate)

	// 检查当前机器人账户余额
	b := &model.BRC20TokenBalance{}
	balanceInfo, err := b.GetByTickerAndAddress(ticker, curRobot.Account)
	if err != nil {
		logrus.Error("[List] get ticker info: ", err)
		return err
	}
	brc20Balance, _ := strconv.ParseInt(balanceInfo.OverallBalance, 10, 64)
	if brc20Balance == 0 {
		logrus.Infof("[List] insufficient balance, account: %s, token: %s, balance: %s", curRobot.Account, ticker, balanceInfo.OverallBalance)
		return nil
	}
	logrus.Infof("[List] current robot: %s, token: %s, brc20 balance: %d", curRobot.Account, ticker, brc20Balance)
	// 随机产生挂单数量
	randAmount := 1 + rand.Int63n(listAmount)
	var (
		totalPrice *big.Float
		listRecord *model.ListRecord
	)
	if brc20Balance < randAmount {
		logrus.Infof("brc20 balance(%v) < rand amount(%v)", brc20Balance, randAmount)
		return nil
	}
	logrus.Info("[List] rand amount: ", randAmount)
	// 挂单中心账户
	center := platform.GetMnemonic()
	centerAccount := platform.Mnemonic2Bench32([]byte(center))
	centerPubKey, err := utils.GetPubkeyFromAddress(centerAccount)
	if err != nil {
		logrus.Error("[List] get pubKey from address: ", err)
		return err
	}
	logrus.Infof("[List] new center account: %v, pubKey: %v", centerAccount, centerPubKey)
	// 订单总价
	totalPrice = big.NewFloat(unitPrice * float64(randAmount))
	listRecord = &model.ListRecord{
		Ticker:         ticker,
		User:           curRobot.Account,
		Price:          totalPrice.String(),
		Amount:         strconv.Itoa(int(randAmount)),
		CenterMnemonic: center,
		CenterUser:     centerAccount,
		State:          constant.ListWaiting,
	}

	tx, err := db.RemoteMaster().Begin()
	if err != nil {
		logrus.Error("[List] remoteDB transaction: ", err)
		return err
	}

	lastInsertId, err := listRecord.InsertToDB()
	if err != nil {
		logrus.Error("[List] insert to db: ", err)
		tx.Rollback()
		return err
	}

	// 3. 转账
	fee := big.NewInt(21_000_000) // 21FRA
	resp, err := utils.SendTx(strconv.Itoa(int(brc20Balance-randAmount)), curRobot.PrivateKey, centerPubKey, centerPubKey, listRecord.Amount, ticker, fee.String(), constant.BRC20_OP_TRANSFER)
	if err != nil {
		logrus.Errorf("[List] send tx: %v, robot: %v", err, curRobot.Account)
		tx.Rollback()
		return err
	}

	var result RpcResult
	_ = json.Unmarshal([]byte(resp), &result)

	// 4. 确认转账
	listRecordTemp := &model.ListRecord{Base: model.Base{Id: uint64(lastInsertId)}, User: curRobot.Account}
	if err = listRecordTemp.ConfirmList(); err != nil {
		logrus.Error("[List] confirm list: ", err)
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		logrus.Error("[List] tx commit: ", err)
		tx.Rollback()
		return err
	}

	logrus.Infof("[List] add list ok, seller: %v, tx: %v, listId: %v, token: %v, amount: %v, unitPrice: %v, totalPrice: %v",
		curRobot.Account, result.Result.Hash, lastInsertId, ticker, listRecord.Amount, unitPrice, totalPrice.String())

	return nil
}

func buy(floorPrice string, firstRobotID int64, robotCount int64, ticker string) error {
	logrus.Info("[Buy] current floor price ", floorPrice)
	decFloorPrice, _, err := decimal.NewDecimalFromString(floorPrice)
	if err != nil {
		return err
	}

	// S:L:L:R means Latest Buy Robot
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:B:R").Int64()
	if err != nil {
		logrus.Error("[Buy] redis: ", err)
		return err
	}
	// 1. 获取当前robot
	r := &model.Robot{}
	curRobot, err := r.GetRobotBuyById(uint64(latestRobotId))
	if err != nil {
		logrus.Error("[Buy] get buyRobotId: ", err)
		return err
	}

	defer func() {
		db.MRedis().Set(context.Background(), "S:L:B:R", (latestRobotId+1)%robotCount+firstRobotID, time.Duration(0))
	}()

	// 2. 获取机器人订单
	l := &model.ListRecord{}
	records, err := l.GetRobotListRecord(ticker)
	if err != nil {
		logrus.Error("[Buy] get robot list records: ", err)
		return err
	}

	// 3. 获取当前机器人的fra余额
	balance := utils.GetFraBalance(curRobot.PrivateKey)
	logrus.Infof("[Buy] robot: %s, FRA balance: %d, lists count: %d", curRobot.Account, balance, len(records))

	// 4. 转账并购买
	for _, rec := range records {
		centerAccount := platform.Mnemonic2Bench32([]byte(rec.CenterMnemonic))
		centerPubKey, err := utils.GetPubkeyFromAddress(centerAccount)
		if err != nil {
			logrus.Error("[Buy] get pub key from address: ", err)
			return err
		}

		// 订单总价
		p, _ := new(big.Float).SetString(rec.Price)
		decRecPrice, _, err := decimal.NewDecimalFromString(new(big.Float).Mul(p, big.NewFloat(1_000_000)).String())
		if err != nil {
			return err
		}
		// 订单token数
		recAmount, _ := new(big.Int).SetString(rec.Amount, 10)
		expectPrice := new(big.Int).Mul(recAmount, decFloorPrice.Value)
		// 理论总价
		decExpectPrice, _, err := decimal.NewDecimalFromString(expectPrice.String())
		if err != nil {
			return err
		}
		if decRecPrice.Value.Cmp(decExpectPrice.Value) >= 0 {
			// 价格大于地板价
			logrus.Infof("[Buy] listPrice(%v) >= floorPrice(%v)", decRecPrice.String(), expectPrice.String())
			continue
		}
		if (balance - decRecPrice.Value.Uint64() - constant.TxMinFee) < 0 {
			// 余额不足
			logrus.Info("[Buy] insufficient FRA balance")
			continue
		}
		logrus.Infof("[Buy] listId %v, list amount: %v, list totalPrice: %v, floor totalPrice: %v", rec.Id, rec.Amount, decRecPrice.String(), decExpectPrice.String())
		// 给中心化账户打钱
		resp, err := utils.Transfer(curRobot.PrivateKey, centerPubKey, decRecPrice.String())
		if err != nil {
			logrus.Error("[Buy] transfer: %v, account: %v", err, curRobot.Account)
			return err
		}
		var result1 RpcResult
		if err = json.Unmarshal([]byte(resp), &result1); err != nil {
			logrus.Error("[Buy] unmarshal transfer result: ", err)
			return err
		}

		logrus.Infof("[Buy] transfer to center hash: %s", result1.Result.Hash)

		// 更改挂单状态
		listRecord := &model.ListRecord{Base: model.Base{Id: rec.Id}, ToUser: curRobot.Account}
		tx, err := db.RemoteMaster().Begin()
		if err != nil {
			return err
		}
		err = listRecord.Finished()
		if err != nil {
			return err
		}

		// 需要中心化账户把brc20 token打给购买者, 并且将fra转给上架者
		toPubKey, err := utils.GetPubkeyFromAddress(curRobot.Account)
		if err != nil {
			return err
		}
		receiver, err := utils.GetPubkeyFromAddress(rec.User)
		if err != nil {
			return err
		}

		recPrivateKey := platform.Mnemonic2PrivateKey([]byte(rec.CenterMnemonic))
		if len(recPrivateKey) == 0 {
			return fmt.Errorf("get private key from recorde mnemonic")
		}

		time.Sleep(time.Second * 20) // 等20秒,为了确保交易已上链

		resp, err = utils.SendTx("0", recPrivateKey, receiver, toPubKey, rec.Amount, rec.Ticker, decRecPrice.String(), constant.BRC20_OP_TRANSFER)
		if err != nil {
			logrus.Error("[Buy] send tx error: ", err)
			return err
		}

		var result2 RpcResult
		if err = json.Unmarshal([]byte(resp), &result2); err != nil {
			logrus.Error("[Buy] unmarshal send result: ", err)
			return err
		}

		logrus.Infof("[Buy] send brc20 hash: %s ", result2.Result.Hash)

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			logrus.Error("[Buy] commit error: ", err)
			return err
		}

		logrus.Infof("[Buy] buy ok, buyer: %v, listId: %d, seller: %v, token: %v, amount: %v, totalPrice: %v", curRobot.Account, listRecord.Id, rec.User, rec.Ticker, rec.Amount, rec.Price)

		break
	}

	return nil
}
