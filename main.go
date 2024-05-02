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
	"time"

	"github.com/sirupsen/logrus"
)

func initRobot() error {
	r := &model.Robot{}
	ok, err := r.CreateBatch()
	if err != nil {
		return err
	}
	// 表明其是第一次创建robot, 需要转钱
	if ok {
		// 通过发空投的账户给机器人打钱
		logrus.Info("send fra to robots")
		_, err = utils.SendRobotBatch(os.Getenv(constant.AIRDROP_MNEMONIC))
		return err
	}
	return nil
}

func init() {
	// S:L:M:R means Latest Mint Robot
	//db.MRedis().SetNX(context.Background(), "S:L:M:R", 1, time.Duration(0))
	//db.MRedis().SetNX(context.Background(), "S:L:L:R", 1, time.Duration(0))
	//db.MRedis().SetNX(context.Background(), "S:L:B:R", 1, time.Duration(0))
	err := initRobot()
	if err != nil {
		logrus.Error(err)
		panic(err)
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

func addList() error {
	// S:L:L:R means Latest List Robot
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:L:R").Int64()
	if err != nil {
		return err
	}
	defer func() {
		db.MRedis().Set(context.Background(), "S:L:L:R", latestRobotId%200+1, time.Duration(0))
	}()

	// 1. 获取当前robot
	r := &model.Robot{}
	curRobot, err := r.GetById(uint64(latestRobotId))
	if err != nil {
		return err
	}

	ticker := os.Getenv(constant.ROBOT_TICK)

	b := &model.BRC20TokenBalance{}
	balance, err := b.GetByTickerAndAddress(ticker, curRobot.Account)
	if err != nil {
		return err
	}

	// 2. 插入上架信息
	// 价格FRA 1000 + Rand(50, 100)
	tx, err := db.Master().Begin()
	if err != nil {
		return err
	}
	price := fmt.Sprintf("%d", 1050+rand.Intn(51))
	amount := balance.OverallBalance
	if amount == "" {
		amount = "0"
	}
	// center 挂单中心账户
	center := platform.GetMnemonic()
	centerAccount := platform.Mnemonic2Bench32([]byte(center))
	centerPubkey, err := utils.GetPubkeyFromAddress(centerAccount)
	if err != nil {
		return err
	}
	listRecord := &model.ListRecord{
		Ticker:         ticker,
		User:           curRobot.Account,
		Price:          price,
		Amount:         amount,
		CenterMnemonic: center,
		State:          constant.ListWaiting,
	}
	lastInsertId, err := listRecord.InsertToDB()
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. 转账
	hash, err := utils.SendTx(curRobot.Mnemonic, centerPubkey, centerPubkey, amount, os.Getenv(constant.ROBOT_TICK), price, constant.BRC20_OP_TRANSFER)
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
	return tx.Commit()
}

func buy() error {
	logrus.Info("buy")
	// S:L:L:R means Latest Buy Robot
	latestRobotId, err := db.MRedis().Get(context.Background(), "S:L:B:R").Int64()
	if err != nil {
		return err
	}
	defer func() {
		db.MRedis().Set(context.Background(), "S:L:B:R", latestRobotId%200+1, time.Duration(0))
	}()

	// 1. 获取当前robot
	r := &model.Robot{}
	curRobot, err := r.GetById(uint64(latestRobotId))
	if err != nil {
		return err
	}

	// 2. 获取机器人订单
	l := &model.ListRecord{}
	records, err := l.GetRobotListRecord()
	if err != nil {
		return err
	}

	// 3. 获取当前机器人的fra余额
	balance := utils.GetFraBalance(curRobot.Mnemonic)
	logrus.Infof("%d %s buy, balance %d, records len %d", latestRobotId, curRobot.Account, balance, len(records))

	// 4. 转账 并 购买
	for _, v := range records {

		centerAccount := platform.Mnemonic2Bench32([]byte(v.CenterMnemonic))
		centerPubkey, err := utils.GetPubkeyFromAddress(centerAccount)
		if err != nil {
			return err
		}
		price, _, err := decimal.NewDecimalFromString(v.Price)
		if err != nil {
			return err
		}
		logrus.Infof("listId %d, price %d", v.Id, price.Value.Int64())
		// 如果price大于当前balance-fee
		if price.Value.Cmp(big.NewInt(int64(balance-constant.TX_MIN_FEE))) >= 0 {
			continue
		}
		// 给中心化账户打钱
		hash, err := utils.Transfer(curRobot.Mnemonic, centerPubkey, v.Price)
		if err != nil {
			return err
		}
		logrus.Infof("buy transfer to center hash: %s", hash)
		// 更改挂单状态
		listRecord := &model.ListRecord{Base: model.Base{Id: v.Id}, ToUser: curRobot.Account}
		tx, err := db.Master().Begin()
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

		receiver, err := utils.GetPubkeyFromAddress(v.User)
		if err != nil {
			return err
		}

		hash, err = utils.SendTx(v.CenterMnemonic, receiver, toPubkey, v.Amount, v.Ticker, v.Price, constant.BRC20_OP_TRANSFER)
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

func main() {
	//mintTicker := time.NewTicker(60 * time.Second)
	//addListTicker := time.NewTicker(70 * time.Second)
	//buyTicker := time.NewTicker(70 * time.Second)
	//defer func() {
	//	mintTicker.Stop()
	//	addListTicker.Stop()
	//	buyTicker.Stop()
	//}()
	//for {
	//	select {
	//	case <-mintTicker.C:
	//		err := mint()
	//		if err != nil {
	//			utils.GetLogger().Errorf("mint tick err:%v", err)
	//			continue
	//		}
	//	case <-addListTicker.C:
	//		err := addList()
	//		if err != nil {
	//			utils.GetLogger().Errorf("list tick err:%v", err)
	//			continue
	//		}
	//	case <-buyTicker.C:
	//		err := buy()
	//		if err != nil {
	//			utils.GetLogger().Errorf("buy tick err:%v", err)
	//			continue
	//		}
	//	default:
	//		time.Sleep(10 * time.Millisecond)
	//	}
	//}
	fmt.Println("ok")
}
