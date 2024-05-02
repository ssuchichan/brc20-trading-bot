# brc20-trading-bot

## 依赖
* Golang
* Rust
* PostgreSQL
* Redis

## 环境变量
```bash
# 本机数据库URL
DATABASE_URL=postgres://postgres:csq123456@localhost/postgres
# 远端（交易所、交易平台）数据库URL
REMOTE_DATABASE_URL=postgres://hello:world123456@localhost/postgres
# Redis URL
REDISURL=redis://localhost:6379/0
# 链接时查找库的目录,(加入libplatform.so所在的路径)
LIBRARY_PATH=
# 运行时查找动态库的目录,(加入libplatform.so所在的路径)
LD_LIBRARY_PATH=
# Findora Node RPC
ENDPOINT=https://prod-testnet.prod.findora.org
# Findora node RPC inner api port
PLAT_INNER_PORT=8668
# Findora node RPC tendermint port
PLAT_API_PORT=26657
# 总钱包助记词，用于给生成的地址转币
AIRDROP_MNEMONIC=
# 铭文代币符号
ROBOT_TICK=neo
# 地板价，精确到最小单位（10的6次方）
FLOOR_PRICES="100000000,150000000,200000000,250000000,300000000"
# 指定地板价开始的位置
PRICE_START_INDEX=0
# 购买间隔，单位秒
BUY_INTERVAL=120
# 挂单间隔，单位秒
LIST_INTERVAL=120
# 挂单维持的铭文总量
LIST_LIMIT=10000
```

## 编译运行

### 编译依赖库
```
cd brc20-trading-bot/platform
cargo build --release
```
`target/release`会生成`libplatform.so`

### 编译可执行文件
```
cd brc20-trading-bot
go build
```
