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

const redisNil = "redis: nil"

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

	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:L:R").Int64()
	if err != nil {
		if err.Error() == redisNil {
			db.MRedis().Set(context.Background(), "S:L:L:R", firstRobotListID, time.Duration(0))
		} else {
			logrus.Fatal("S:L:L:R: ", err)
		}
	}
	if latestRobotId == 0 {
		db.MRedis().Set(context.Background(), "S:L:L:R", firstRobotListID, time.Duration(0))
	}

	latestRobotId, err = db.MRedis().Get(context.Background(), "S:L:B:R").Int64()
	if err != nil {
		if err.Error() == redisNil {
			db.MRedis().Set(context.Background(), "S:L:B:R", firstRobotBuyID, time.Duration(0))
		} else {
			logrus.Fatal("S:L:B:R: ", err)
		}
	}
	if latestRobotId == 0 {
		db.MRedis().Set(context.Background(), "S:L:B:R", firstRobotBuyID, time.Duration(0))
	}

	// S:L:M:R means Latest Mint Robot
	//db.MRedis().SetNX(context.Background(), "S:L:M:R", 1, time.Duration(0))
	// Least List Robot
	//db.MRedis().SetNX(context.Background(), "S:L:L:R", firstRobotListID, time.Duration(0))
	// Least Buy Robot
	//db.MRedis().SetNX(context.Background(), "S:L:B:R", firstRobotBuyID, time.Duration(0))
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
	if err != nil {
		logrus.Fatal("ROBOT_TICK: ", err)
	}
	listLimit, err = strconv.ParseInt(os.Getenv("LIST_LIMIT"), 10, 64)
	if err != nil {
		logrus.Fatal("LIST_LIMIT: ", err)
	}
	priceIndex, err = strconv.ParseInt(os.Getenv("PRICE_START_INDEX"), 10, 64)
	if err != nil {
		logrus.Fatal("PRICE_START_INDEX: ", err)
	}
	listInterval, err = strconv.ParseInt(os.Getenv("LIST_INTERVAL"), 10, 64)
	if err != nil {
		logrus.Fatal("LIST_INTERVAL: ", err)
	}
	buyInterval, err = strconv.ParseInt(os.Getenv("BUY_INTERVAL"), 10, 64)
	if err != nil {
		logrus.Fatal("BUY_INTERVAL: ", err)
	}
	priceUpdateInterval, err = strconv.ParseInt(os.Getenv("FLOOR_PRICE_UPDATE_INTERVAL"), 10, 64)
	if err != nil {
		logrus.Fatal("FLOOR_PRICE_UPDATE_INTERVAL: ", err)
	}
	listAmount, err = strconv.ParseInt(os.Getenv("LIST_AMOUNT"), 10, 64)
	if err != nil {
		logrus.Fatal("LIST_AMOUNT: ", err)
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
	rList := &model.Robot{}
	firstListRobotID, err := rList.GetFirstRobotList()
	if err != nil {
		logrus.Fatal("Get the first listRobotId: ", err)
	}
	robotListCount, _ := rList.GetListRobotCount()
	if robotListCount == 0 {
		logrus.Fatal("The listRobot count is 0")
	}
	logrus.Infof("The first list robot id: %v, list robot count: %v", firstListRobotID, robotListCount)
	rBuy := &model.Robot{}
	firstBuyRobotID, err := rBuy.GetFirstRobotBuy()
	if err != nil {
		logrus.Fatal("Get the first listRobotId: ", err)
	}
	robotBuyCount, _ := rBuy.GetBuyRobotCount()
	if robotBuyCount == 0 {
		logrus.Fatal("The buyRobot count is 0")
	}
	logrus.Infof("The first buy robot id: %v, buy robot count: %v", firstBuyRobotID, robotBuyCount)
	for {
		select {
		//case <-mintTicker.C:
		//	err := mint()
		//	if err != nil {
		//		utils.GetLogger().Errorf("mint tick err:%v", err)
		//		continue
		//	}
		case <-addListTicker.C:
			curFloorPrice := floorPrices[priceIndex]
			err = addList(curFloorPrice, listLimit, listAmount, int64(firstListRobotID), int64(robotListCount), token)
			if err != nil {
				logrus.Errorf("[List] list err: %v", err)
			}
		case <-buyTicker.C:
			curFloorPrice := floorPrices[priceIndex]
			err = buy(curFloorPrice, int64(firstBuyRobotID), int64(robotBuyCount), token)
			if err != nil {
				logrus.Errorf("[Buy] buy err: %v", err)
			}
		case <-priceTicker.C:
			if priceIndex+1 == int64(len(prices)) {
				logrus.Info("[FloorPrice] Reached the last floor price, exit.")
				return
			}
			priceIndex += 1
			logrus.Info("[FloorPrice] update floor price to: ", floorPrices[priceIndex])
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
	// S:L:L:R means Latest List Robot
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:L:R").Int64()
	if err != nil {
		logrus.Error("[List] redis: ", err)
		return err
	}
	if latestRobotId == 0 {
		return fmt.Errorf("invalid robot id: 0")
	}
	logrus.Info("[List] current list robot id: ", latestRobotId)
	// 1. 获取当前robot
	rb := &model.Robot{}
	curRobot, err := rb.GetRobotListById(uint64(latestRobotId))
	if err != nil {
		logrus.Error("[List] get listRobot id: ", err)
		return err
	}

	defer func() {
		nextID := latestRobotId + 1
		if nextID > (firstRobotID + robotCount - 1) {
			nextID = firstRobotID
		}
		db.MRedis().Set(context.Background(), "S:L:L:R", nextID, time.Duration(0))
	}()

	// 当前list总量
	rec := &model.ListRecord{}
	fp, _ := strconv.ParseFloat(floorPrice, 64)
	totalList, err := rec.SumListAmount(ticker, 1.5*fp)
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
	unitPrice := fp * rate // 实际单价
	logrus.Infof("[List] floorPirce: %v, unitPrice: %v, rate: %v", floorPrice, fmt.Sprintf("%.2f", unitPrice), rate)

	// 检查当前机器人账户余额
	b := &model.BRC20TokenBalance{}
	balanceInfo, err := b.GetByTickerAndAddress(ticker, curRobot.Account)
	if err != nil {
		logrus.Error("[List] get ticker info: ", err)
		return err
	}
	brc20Balance, _ := strconv.ParseInt(balanceInfo.OverallBalance, 10, 64)
	if brc20Balance == 0 {
		logrus.Infof("[List] insufficient BRC20 balance, account: %s, token: %s, balance: %v", curRobot.Account, ticker, brc20Balance)
		return nil
	}
	logrus.Infof("[List] current robot: %s, token: %s, brc20 balance: %d", curRobot.Account, ticker, brc20Balance)
	// 随机产生挂单数量
	randAmount := 1 + rand.Int63n(listAmount)
	var (
		totalPrice string
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
	if len(centerAccount) == 0 {
		return fmt.Errorf("[List] centerAccount is 0")
	}
	centerPubKey, err := utils.GetPubkeyFromAddress(centerAccount)
	if len(centerPubKey) == 0 {
		return fmt.Errorf("[List] centerPubKey is 0")
	}
	if err != nil {
		logrus.Error("[List] get pubKey from address: ", err)
		return err
	}
	logrus.Infof("[List] new center account: %v, pubKey: %v", centerAccount, centerPubKey)
	// 订单总价
	totalPrice = fmt.Sprintf("%.2f", unitPrice*float64(randAmount))
	listRecord = &model.ListRecord{
		Ticker:         ticker,
		User:           curRobot.Account,
		Price:          totalPrice,
		Amount:         strconv.Itoa(int(randAmount)),
		CenterMnemonic: center,
		CenterUser:     centerAccount,
		State:          constant.ListWaiting,
	}

	tx, err := db.RemoteMaster().Begin()
	if err != nil {
		tx.Rollback()
		logrus.Error("[List] remoteDB transaction: ", err)
		return err
	}

	lastInsertId, err := listRecord.InsertToDB()
	if err != nil {
		tx.Rollback()
		logrus.Error("[List] insert to db: ", err)
		return err
	}

	// 3. 转账
	remain := int(brc20Balance - randAmount)
	fee := strconv.Itoa(14_000_000) // 14 FRA
	resp, err := utils.SendTx(strconv.Itoa(remain), curRobot.PrivateKey, centerPubKey, centerPubKey, listRecord.Amount, ticker, fee, constant.BRC20_OP_TRANSFER)
	if err != nil {
		tx.Rollback()
		logrus.Errorf("[List] send tx: %v, robot: %v", err, curRobot.Account)
		return err
	}

	var result RpcResult
	if err = json.Unmarshal([]byte(resp), &result); err != nil {
		tx.Rollback()
		logrus.Errorf("[List] unmarshal: %v", err)
		return err
	}

	// 4. 确认转账
	listRecordTemp := &model.ListRecord{Base: model.Base{Id: uint64(lastInsertId)}, User: curRobot.Account}
	if err = listRecordTemp.ConfirmList(); err != nil {
		tx.Rollback()
		logrus.Error("[List] confirm list: ", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logrus.Error("[List] commit transaction: ", err)
		return err
	}

	logrus.Infof("[List] add list ok, seller: %v, tx: %v, listId: %v, token: %v, amount: %v, unitPrice: %v, totalPrice: %v",
		curRobot.Account, result.Result.Hash, lastInsertId, ticker, listRecord.Amount, fmt.Sprintf("%.2f", unitPrice), totalPrice)

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
	if latestRobotId == 0 {
		return fmt.Errorf("invalid robot id: 0")
	}
	logrus.Info("[Buy] current buy robot id: ", latestRobotId)

	// 1. 获取当前robot
	r := &model.Robot{}
	curRobot, err := r.GetRobotBuyById(uint64(latestRobotId))
	if err != nil {
		logrus.Error("[Buy] get buyRobotId: ", err)
		return err
	}

	defer func() {
		nextID := latestRobotId + 1
		if nextID > (firstRobotID + robotCount - 1) {
			nextID = firstRobotID
		}
		db.MRedis().Set(context.Background(), "S:L:B:R", nextID, time.Duration(0))
	}()

	// 2. 获取订单
	l := &model.ListRecord{}
	records, err := l.GetListRecord(ticker)
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
		decRecPrice, _, err := decimal.NewDecimalFromString(rec.Price)
		if err != nil {
			return err
		}
		// 订单token数
		recAmount, _ := new(big.Int).SetString(rec.Amount, 10)
		// 理论总价
		expectPrice := new(big.Int).Mul(recAmount, decFloorPrice.Value)
		if decRecPrice.Value.Cmp(expectPrice) >= 0 {
			// 价格大于地板价
			logrus.Infof("[Buy] listId: %v, listAmount: %v, listPrice: %v (with precision) >= expectPrice: %v (with precision)", rec.Id, rec.Amount, decRecPrice.Value.String(), expectPrice.String())
			continue
		}
		if (balance - decRecPrice.Value.Uint64() - constant.TxMinFee) < 0 {
			// 余额不足
			logrus.Info("[Buy] insufficient FRA balance")
			continue
		}
		logrus.Infof("[Buy] listId: %v, listAmount: %v, listPrice: %v (with precision), expectPrice: %v (with precision)", rec.Id, rec.Amount, decRecPrice.Value.String(), expectPrice.String())
		// 给中心化账户打钱
		resp, err := utils.Transfer(curRobot.PrivateKey, centerPubKey, decRecPrice.Value.String())
		if err != nil {
			logrus.Errorf("[Buy] transfer: %v, account: %v", err, curRobot.Account)
			return err
		}

		time.Sleep(time.Second * 20) // 等20秒,为了确保交易已上链

		var result1 RpcResult
		if err = json.Unmarshal([]byte(resp), &result1); err != nil {
			logrus.Error("[Buy] unmarshal transfer result: ", err)
			return err
		}
		if result1.Result.Code > 0 {
			return fmt.Errorf("[Buy] transfer to center failed")
		}

		logrus.Infof("[Buy] transfer to center hash: %s", result1.Result.Hash)

		// 更改挂单状态
		listRecord := &model.ListRecord{Base: model.Base{Id: rec.Id}, ToUser: curRobot.Account}
		tx, err := db.RemoteMaster().Begin()
		if err != nil {
			tx.Rollback()
			logrus.Error("[Buy] begin db transaction: ", err)
			return err
		}
		err = listRecord.Finished()
		if err != nil {
			tx.Rollback()
			logrus.Error("[Buy] finish record list: ", err)
			return err
		}

		// 需要中心化账户把brc20 token打给购买者, 并且将fra转给上架者
		// 购买者
		toPubKey, err := utils.GetPubkeyFromAddress(curRobot.Account)
		if err != nil {
			tx.Rollback()
			logrus.Error("[Buy] get pub key from address: ", err)
			return err
		}
		// 上架者/挂单者
		fraReceiver, err := utils.GetPubkeyFromAddress(rec.User)
		if err != nil {
			tx.Rollback()
			logrus.Error("[Buy] get pub key from address: ", err)
			return err
		}
		centerPrivateKey := platform.Mnemonic2PrivateKey([]byte(rec.CenterMnemonic))
		if len(centerPrivateKey) == 0 {
			tx.Rollback()
			return fmt.Errorf("[Buy] center privateKey is 0")
		}

		resp, err = utils.SendTx("0", centerPrivateKey, fraReceiver, toPubKey, rec.Amount, rec.Ticker, decRecPrice.Value.String(), constant.BRC20_OP_TRANSFER)
		if err != nil {
			tx.Rollback()
			logrus.Error("[Buy] send tx error: ", err)
			return err
		}

		var result2 RpcResult
		if err = json.Unmarshal([]byte(resp), &result2); err != nil {
			tx.Rollback()
			logrus.Error("[Buy] unmarshal send result: ", err)
			return err
		}

		logrus.Infof("[Buy] send brc20 hash: %s ", result2.Result.Hash)

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			logrus.Error("[Buy] commit error: ", err)
			return err
		}

		logrus.Infof("[Buy] buy ok, buyer: %v, listId: %d, seller: %v, token: %v, amount: %v, totalPrice: %v", curRobot.Account, listRecord.Id, rec.User, rec.Ticker, rec.Amount, rec.Price)

		break
	}

	return nil
}
