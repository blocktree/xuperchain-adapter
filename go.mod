module github.com/blocktree/xuperchain-adapter

go 1.13

require (
	github.com/astaxie/beego v1.12.1
	github.com/blocktree/go-owcdrivers v1.2.0
	github.com/blocktree/go-owcrypt v1.1.2
	github.com/blocktree/openwallet v1.5.4
	github.com/blocktree/openwallet/v2 v2.0.4
	github.com/blocktree/quorum-adapter v1.2.1 // indirect
	github.com/ethereum/go-ethereum v1.9.9
	github.com/golang/protobuf v1.3.2
	github.com/imroc/req v0.3.0
	github.com/shopspring/decimal v1.2.0
	github.com/tidwall/gjson v1.6.0
	github.com/xuperchain/xuper-sdk-go v0.0.0-20200407074302-fd8273561271
	github.com/xuperchain/xuperchain v0.0.0-20200312070156-efa72c51cef3
	google.golang.org/grpc v1.24.0
)

replace github.com/blocktree/openwallet/v2 => ../../openwallet
