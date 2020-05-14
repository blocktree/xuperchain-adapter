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
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/shopspring/decimal"
)

const (
	maxExtractingSize = 10 //并发的扫描线程数
)

//BlockScanner 区块链扫描器
type BlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	IsScanMemPool        bool           //是否扫描交易池
	RescanLastBlockCount uint64         //重扫上N个区块数量
}

//ExtractResult 扫描完成的提取结果
type ExtractResult struct {
	extractData map[string]*openwallet.TxExtractData
	TxID        string
	BlockHeight uint64
	Success     bool
}

//SaveResult 保存结果
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

//NewBlockScanner 创建区块链扫描器
func NewBlockScanner(wm *WalletManager) *BlockScanner {
	bs := BlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm
	bs.RescanLastBlockCount = 0

	//设置扫描任务
	//bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//GetBalanceByAddress
func (bs *BlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

    balanceArray := make([]*openwallet.Balance, 0)
	for _, addr := range address {
		balance, err := bs.wm.RPC.GetBalance(addr)
		if err != nil {
			continue
		}
		cb, _ := decimal.NewFromString(balance.Balance)
		cb = cb.Shift(-bs.wm.Decimal())
		b := &openwallet.Balance{
			Symbol:           bs.wm.Symbol(),
			Address:          addr,
			ConfirmBalance:   cb.String(),
			UnconfirmBalance: "",
			Balance:          cb.String(),
		}
		balanceArray = append(balanceArray, b)
	}

	return balanceArray, nil

}