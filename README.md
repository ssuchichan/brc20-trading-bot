# brc20-trading-bot

## 依赖
* Golang
* Rust
* PostgreSQL
* Redis

## 环境变量
```
DATABSE_URL=postgres://postgres:csq123456@localhost/postgres
REMOTE_DATABSE_URL=postgres://hello:world123456@localhost/postgres
REDISURL=redis://localhost:6379/0
LIBRARY_PATH=libplatform.so所在的路径
LD_LIBRARY_PATH=libplatform.so所在的路径
ENDPOINT=https://prod-testnet.prod.findora.org
PLAT_INNER_PORT=8668
PLAT_API_PORT=26657
AIRDROP_MNEMONIC=助记词，用于给生成的地址转币
ROBOT_TICK=
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
