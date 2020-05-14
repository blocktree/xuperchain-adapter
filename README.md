# xuperchain-adapter

xuperchain-adapter适配了openwallet.AssetsAdapter接口，给应用提供了底层的区块链协议支持。

## 如何测试

openwtester包下的测试用例已经集成了openwallet钱包体系，创建conf文件，XUPER.ini文件，编辑如下内容：

```ini

# RPC api url
serverAPI = "127.0.0.1:37101"
# chain name
chainName = "xuper"

```

## 主链启动

nohup ./xchain --vm ixvm &


## 编译合约

GOOS=js GOARCH=wasm go build XXX.go