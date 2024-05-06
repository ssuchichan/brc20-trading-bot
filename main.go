package main

import (
	"brc20-trading-bot/constant"
	"brc20-trading-bot/db"
	"brc20-trading-bot/decimal"
	"brc20-trading-bot/model"
	"brc20-trading-bot/platform"
	"brc20-trading-bot/utils"
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func initRobot() error {
	logrus.Info("Init accounts...")
	r := &model.Robot{}
	ok, err := r.CreateBatch()
	if err != nil {
		logrus.Info("Init robot accounts...ok")
		return err
	}

	if ok {
		// 第一次创建robot, 需要转钱
		// 通过发空投的账户给机器人打钱
		logrus.Info("Sending FRA...")
		_, err = utils.SendRobotBatch(os.Getenv(constant.AIRDROP_MNEMONIC))
		return err
	}
	logrus.Info("Init accounts...ok")
	return nil
}

func init() {
	// S:L:M:R means Latest Mint Robot
	//db.MRedis().SetNX(context.Background(), "S:L:M:R", 1, time.Duration(0))

	var (
		firstRobotListID uint64
		firstRobotBuyID  uint64
		err              error
	)

	// Least List Robot
	db.MRedis().SetNX(context.Background(), "S:L:L:R", firstRobotListID, time.Duration(0))
	// Least Buy Robot
	db.MRedis().SetNX(context.Background(), "S:L:B:R", firstRobotBuyID, time.Duration(0))

	if err = initRobot(); err != nil {
		logrus.Error(err)
		panic(err)
	}

	r := &model.Robot{}
	firstRobotListID, err = r.GetFirstRobotList()
	firstRobotBuyID, err = r.GetFirstRobotBuy()
	logrus.Infof("The first listRobotId: %d, the first buyRobotId: %d", firstRobotListID, firstRobotBuyID)
}

func main() {
	var (
		floorPrices         []int64
		priceIndex          int64
		listInterval        int64
		buyInterval         int64
		priceUpdateInterval int64
		listLimit           int64
		err                 error
	)
	floorPricesStr := os.Getenv("FLOOR_PRICES")
	prices := strings.Split(floorPricesStr, ",")
	for i := 0; i < len(prices); i++ {
		p, _ := strconv.ParseInt(prices[i], 10, 64)
		floorPrices = append(floorPrices, p*1_000_000)
	}
	listLimit, err = strconv.ParseInt(os.Getenv("LIST_LIMIT"), 10, 64)
	priceIndex, err = strconv.ParseInt(os.Getenv("PRICE_START_INDEX"), 10, 64)
	listInterval, err = strconv.ParseInt(os.Getenv("LIST_INTERVAL"), 10, 64)
	buyInterval, err = strconv.ParseInt(os.Getenv("BUY_INTERVAL"), 10, 64)
	priceUpdateInterval, err = strconv.ParseInt(os.Getenv("FLOOR_PRICE_UPDATE_INTERVAL"), 10, 64)
	if err != nil {
		panic(err)
	}
	logrus.Info("Floor prices: ", floorPrices)
	logrus.Info("Current floor price: ", floorPrices[priceIndex])
	logrus.Infof("Floor prices updating interval: %ds", priceUpdateInterval)
	logrus.Info("List limit: ", listLimit)
	logrus.Infof("List interval: %ds, buy inteval: %ds", listInterval, buyInterval)

	//mintTicker := time.NewTicker(60 * time.Second)
	addListTicker := time.NewTicker(time.Duration(listInterval) * time.Second)
	buyTicker := time.NewTicker(time.Duration(buyInterval+120) * time.Second)
	priceTicker := time.NewTicker(time.Duration(priceUpdateInterval) * time.Second)
	defer func() {
		//mintTicker.Stop()
		addListTicker.Stop()
		buyTicker.Stop()
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
			curFloorPrice := floorPrices[priceIndex]
			err := addList(curFloorPrice, listLimit)
			if err != nil {
				utils.GetLogger().Errorf("list tick err:%v", err)
				continue
			}
		case <-buyTicker.C:
			curFloorPrice := floorPrices[priceIndex]
			err := buy(curFloorPrice)
			if err != nil {
				utils.GetLogger().Errorf("buy tick err:%v", err)
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

	tick := os.Getenv(constant.ROBOT_TICK)

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

	_, err = utils.SendTx(curRobot.Mnemonic, curPubkey, curPubkey, token.Limit, os.Getenv(constant.ROBOT_TICK), "0", constant.BRC20_OP_MINT)
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

func addList(floorPrice int64, listLimit int64) error {
	// S:L:L:R means Latest List Robot
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:L:R").Int64()
	if err != nil {
		return err
	}

	// 1. 获取当前robot
	r := &model.Robot{}
	curRobot, err := r.GetRobotListById(uint64(latestRobotId))
	if err != nil {
		return err
	}
	robotCount, _ := r.GetListRobotCount()
	if robotCount == 0 {
		return fmt.Errorf("the accounts count is 0")
	}

	defer func() {
		db.MRedis().Set(context.Background(), "S:L:L:R", (latestRobotId+1)%int64(robotCount), time.Duration(0))
	}()

	// 当前list总量
	rec := &model.ListRecord{}
	totalList, err := rec.SumListAmount(r.Account)
	if err != nil {
		return err
	}
	if totalList >= listLimit {
		logrus.Infof("Reach list limit, total listed: %d, list limit: %d", totalList, listLimit)
		return nil
	}
	logrus.Info("Current list amount: %d", totalList)
	ticker := os.Getenv(constant.ROBOT_TICK)
	delta := listLimit - totalList
	tx, err := db.RemoteMaster().Begin()
	if err != nil {
		return err
	}
	var checkRetry int
	for delta > 0 {
		// 检查当前机器人账户余额
		b := &model.BRC20TokenBalance{}
		balanceInfo, err := b.GetByTickerAndAddress(ticker, curRobot.Account)
		if err != nil {
			return err
		}
		amount, _ := strconv.ParseInt(balanceInfo.OverallBalance, 10, 64)
		if amount == 0 {
			if (checkRetry + 1) == int(robotCount) {
				return fmt.Errorf("all accounts are incificient")
			}
			nextRobot, err := curRobot.NextListRobot(robotCount)
			if err != nil {
				return err
			}
			logrus.Infof("%s account is incificient balance, next account: %s", curRobot.Account, nextRobot.Account)
			curRobot = nextRobot
			continue
		}
		price := fmt.Sprintf("%d", floorPrice+floorPrice*int64(rand.Intn(101)/100))
		// center 挂单中心账户
		//center := platform.GetMnemonic()
		//centerAccount := platform.Mnemonic2Bench32([]byte(center))
		center := platform.GeneratePrivateKey()
		centerAccount := platform.PrivateKey2Bech32([]byte(center))
		centerPubKey, err := utils.GetPubkeyFromAddress(centerAccount)
		if err != nil {
			return err
		}
		listRecord := &model.ListRecord{
			Ticker:         ticker,
			User:           curRobot.Account,
			Price:          price,
			Amount:         balanceInfo.OverallBalance,
			CenterMnemonic: center,
			State:          constant.ListWaiting,
		}
		lastInsertId, err := listRecord.InsertToDB()
		if err != nil {
			tx.Rollback()
			return err
		}

		// 3. 转账
		hash, err := utils.SendTx(curRobot.PrivateKey, centerPubKey, centerPubKey, balanceInfo.OverallBalance, os.Getenv(constant.ROBOT_TICK), price, constant.BRC20_OP_TRANSFER)
		if err != nil {
			tx.Rollback()
			return err
		}
		logrus.Infof("list transfer hash: %s", hash)

		// 4. 确认转账
		listRecordTemp := &model.ListRecord{Base: model.Base{Id: uint64(lastInsertId)}, User: curRobot.Account}
		err = listRecordTemp.ConfirmList()
		if err != nil {
			tx.Rollback()
			return err
		}

		delta -= amount
	}

	return tx.Commit()
}

func buy(floorPrice int64) error {
	logrus.Info("buy, current floor price ", floorPrice)
	// S:L:L:R means Latest Buy Robot
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:B:R").Int64()
	if err != nil {
		return err
	}
	// 1. 获取当前robot
	r := &model.Robot{}
	curRobot, err := r.GetRobotBuyById(uint64(latestRobotId))
	if err != nil {
		return err
	}
	robotCount, _ := r.GetBuyRobotCount()
	if robotCount == 0 {
		return fmt.Errorf("the accounts count is 0")
	}
	defer func() {
		db.MRedis().Set(context.Background(), "S:L:B:R", (latestRobotId+1)%int64(robotCount), time.Duration(0))
	}()

	// 2. 获取机器人订单
	l := &model.ListRecord{}
	records, err := l.GetRobotListRecord()
	if err != nil {
		return err
	}

	// 3. 获取当前机器人的fra余额
	balance := utils.GetFraBalance(curRobot.PrivateKey)
	logrus.Infof("%d %s buy, balance %d, records len %d", latestRobotId, curRobot.Account, balance, len(records))

	// 4. 转账并购买
	for _, rec := range records {
		//centerAccount := platform.Mnemonic2Bench32([]byte(v.CenterMnemonic))
		//centerPubKey, err := utils.GetPubkeyFromAddress(centerAccount)
		centerAccount := platform.PrivateKey2Bech32([]byte(rec.CenterMnemonic))
		centerPubKey, err := utils.GetPubkeyFromAddress(centerAccount)
		if err != nil {
			return err
		}
		price, _, err := decimal.NewDecimalFromString(rec.Price)
		if err != nil {
			return err
		}
		logrus.Infof("listId %d, price %d", rec.Id, price.Value.Int64())
		priceDec, _, _ := decimal.NewDecimalFromString(rec.Price)
		amountDec, _, _ := decimal.NewDecimalFromString(rec.Amount)
		payment := new(big.Int).Mul(amountDec.Value, priceDec.Value)
		if price.Value.Cmp(big.NewInt(floorPrice)) > 0 || (balance-payment.Uint64()-constant.TX_MIN_FEE) < 0 {
			// 价格大于地板价
			// 余额不足
			continue
		}
		// 给中心化账户打钱
		hash, err := utils.Transfer(curRobot.PrivateKey, centerPubKey, payment.String())
		if err != nil {
			return err
		}
		logrus.Infof("buy transfer to center hash: %s", hash)
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
		toPubkey, err := utils.GetPubkeyFromAddress(curRobot.Account)
		if err != nil {
			return err
		}

		receiver, err := utils.GetPubkeyFromAddress(rec.User)
		if err != nil {
			return err
		}

		hash, err = utils.SendTx(rec.CenterMnemonic, receiver, toPubkey, rec.Amount, rec.Ticker, rec.Price, constant.BRC20_OP_TRANSFER)
		if err != nil {
			return err
		}
		logrus.Infof("buy send brc20 hash: %s ", hash)

		if err := tx.Commit(); err != nil {
			return err
		}
		break
	}

	return nil
}
