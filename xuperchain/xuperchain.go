/*
 * Copyright 2019 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package xuperchain

import (
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/blocktree/xuperchain-adapter/xuperchain_rpc"
)

//FullName 币种全名
func (wm *WalletManager) FullName() string {
	return "xuperchain"
}

//CurveType 曲线类型
func (wm *WalletManager) CurveType() uint32 {
	return wm.Config.CurveType
}

//Symbol 币种标识
func (wm *WalletManager) Symbol() string {
	return wm.Config.Symbol
}

//小数位精度
func (wm *WalletManager) Decimal() int32 {
	return 8
}

//AddressDecode 地址解析器
func (wm *WalletManager) GetAddressDecoderV2() openwallet.AddressDecoderV2 {
	return wm.AddrDecoder
}

//TransactionDecoder 交易单解析器
func (wm *WalletManager) GetTransactionDecoder() openwallet.TransactionDecoder {
	return wm.TxDecoder
}

//GetBlockScanner 获取区块链
func (wm *WalletManager) GetBlockScanner() openwallet.BlockScanner {
	return wm.BlockScanner
}

//LoadAssetsConfig 加载外部配置
func (wm *WalletManager) LoadAssetsConfig(c config.Configer) error {
	wm.Config.ServerAPI = c.String("serverAPI")
	wm.Config.ChainName = c.String("chainName")
	client := xuperchain_rpc.NewClient(wm.Config.ServerAPI, wm.Config.ChainName)
	wm.RPC = client
	return nil
}

//InitAssetsConfig 初始化默认配置
func (wm *WalletManager) InitAssetsConfig() (config.Configer, error) {
	return config.NewConfigData("ini", []byte(""))
}

//GetAssetsLogger 获取资产账户日志工具
func (wm *WalletManager) GetAssetsLogger() *log.OWLogger {
	return wm.Log
}

//GetSmartContractDecoder 获取智能合约解析器
func (wm *WalletManager) GetSmartContractDecoder() openwallet.SmartContractDecoder {
	return wm.ContractDecoder
}

func (wm *WalletManager) BalanceModelType() openwallet.BalanceModelType {
	return openwallet.BalanceModelTypeAddress
}
