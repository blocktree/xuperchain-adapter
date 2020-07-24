/*
 * Copyright 2018 The openwallet Authors
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

package openwtester

import (
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/v2/common/file"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openw"
	"github.com/blocktree/openwallet/v2/openwallet"
	"path/filepath"
	"testing"
)

////////////////////////// 测试单个扫描器 //////////////////////////

type subscriberSingle struct {
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriberSingle) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Notice("header:", header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriberSingle) BlockExtractDataNotify(sourceKey string, data *openwallet.TxExtractData) error {
	log.Notice("account:", sourceKey)

	for i, input := range data.TxInputs {
		log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
	}

	for i, output := range data.TxOutputs {
		log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
	}

	log.Std.Notice("data.Transaction: %+v", data.Transaction)

	return nil
}

//BlockExtractSmartContractDataNotify 区块提取智能合约交易结果通知
func (sub *subscriberSingle) BlockExtractSmartContractDataNotify(sourceKey string, data *openwallet.SmartContractReceipt) error {

	log.Notice("sourceKey:", sourceKey)
	log.Std.Notice("data.ContractTransaction: %+v", data)

	for i, event := range data.Events {
		log.Std.Notice("data.Events[%d]: %+v", i, event)
	}

	return nil
}

func TestSubscribeAddress_XUPER(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "XUPER"
		addrs = map[string]string{
			"UGbV2vBqMFH4teW7GeaEz19nt7pA5CuT3": "sender",
			"Rvm1AE6rZwLpPFbcBfD7wZxXK3FR6QEXb": "receiver",
		}
	)

	scanTargetFunc := func(target openwallet.ScanTargetParam) openwallet.ScanTargetResult {
		sourceKey, ok := addrs[target.ScanTarget]
		return openwallet.ScanTargetResult{SourceKey: sourceKey, Exist: ok, TargetInfo: nil,}
	}

	scanner := testBlockScanner(symbol)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}
	scanner.SetBlockScanTargetFuncV2(scanTargetFunc)
	scanner.SetRescanBlockHeight(258917)
	scanner.Run()

	<-endRunning
}

func TestBlockScanner_ExtractTransactionAndReceiptData(t *testing.T) {

	var (
		symbol = "XUPER"
		addrs  = make(map[string]openwallet.ScanTargetResult)
		txid   = "8b857276f4b629d0bff671b35268bce7199a11c68dd96f3fa5e7cbef5e5a95d3"
	)

	contract := &openwallet.SmartContract{
		Symbol:   "XUPER",
		Address:  "wasm:artToyContract2",
		Decimals: 8,
	}
	contract.ContractID = openwallet.GenContractID(contract.Symbol, contract.Address)
	contract.SetABI(`[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"from","type":"address"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"string","name":"orderNum","type":"string"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"string","name":"orderNum","type":"string"}],"name":"Burn","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"orderNum","type":"string"}],"name":"issue","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"orderNum","type":"string"}],"name":"burn","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"address","name":"owner","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"allowance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"address","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"totalSupply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":false,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"owner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"AddMerchant","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"RemoveMerchant","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"addMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"removeMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"isMerchant","outputs":[{"internalType":"bool","name":"flag","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"AddArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"}],"name":"RemoveArtToyFromSeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"uint64","name":"index","type":"uint64"},{"indexed":false,"internalType":"string","name":"owner","type":"string"}],"name":"PurchaseArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"string","name":"owner","type":"string"}],"name":"ReceiveArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"ReturnArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"uint64","name":"status","type":"uint64"},{"indexed":false,"internalType":"uint256","name":"drawPrice","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"returnPrice","type":"uint256"}],"name":"SetArtToySeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"RevokeArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"}],"name":"RevokeArtToySeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"orderNumber","type":"string"},{"indexed":false,"internalType":"string","name":"sellerAddr","type":"string"},{"indexed":false,"internalType":"string","name":"buyerAddr","type":"string"},{"indexed":false,"internalType":"uint256","name":"offerAmount","type":"uint256"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"uint64","name":"orderType","type":"uint64"},{"indexed":false,"internalType":"uint64","name":"status","type":"uint64"}],"name":"OrderArtToyExchange","type":"event"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"isArtToySeriesExist","outputs":[{"internalType":"bool","name":"flag","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"}],"name":"newArtToySeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"}],"name":"addArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"}],"name":"removeArtToyFromSeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"}],"name":"revokeArtToySeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"revokeArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"uint256","name":"drawPrice","type":"uint256"},{"internalType":"uint256","name":"returnPrice","type":"uint256"},{"internalType":"uint256","name":"status","type":"uint256"}],"name":"setArtToySeriesInfo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"merchant","type":"string"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"purchaseArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"receiveArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"returnArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"orderType","type":"uint256"}],"name":"orderArtToyExchange","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"}],"name":"getArtToyExchangeOrder","outputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"sellerAddr","type":"string"},{"internalType":"string","name":"buyerAddr","type":"string"},{"internalType":"uint256","name":"offerAmount","type":"uint256"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint64","name":"orderType","type":"uint64"},{"internalType":"uint64","name":"status","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"number","type":"string"}],"name":"dealArtToyExchange","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"}],"name":"cancelArtToyExchangeOrder","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"getArtToyByNumber","outputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint256","name":"merchant","type":"uint256"},{"internalType":"string","name":"owner","type":"string"},{"internalType":"uint64","name":"creatAt","type":"uint64"},{"internalType":"uint64","name":"status","type":"uint64"},{"internalType":"uint64","name":"index","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"merchant","type":"string"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"getArtToySeries","outputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"uint256","name":"merchant","type":"uint256"},{"internalType":"uint64","name":"status","type":"uint64"},{"internalType":"uint256","name":"drawPrice","type":"uint256"},{"internalType":"uint256","name":"returnPrice","type":"uint256"},{"internalType":"uint64","name":"toySize","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	addrs[contract.Address] = openwallet.ScanTargetResult{SourceKey: contract.ContractID, Exist: true, TargetInfo: contract}

	//addrs["0x19a4b5d6ea319a5d5ad1d4cc00a5e2e28cac5ec3"] = openwallet.ScanTargetResult{
	//	SourceKey:  "receiver",
	//	Exist:      true,
	//	TargetInfo: nil,
	//}

	scanTargetFunc := func(target openwallet.ScanTargetParam) openwallet.ScanTargetResult {
		if target.ScanTargetType == openwallet.ScanTargetTypeContractAddress {
			return addrs[target.ScanTarget]
		} else if target.ScanTargetType == openwallet.ScanTargetTypeAccountAddress {
			return addrs[target.ScanTarget]
		}
		return openwallet.ScanTargetResult{SourceKey: "", Exist: false, TargetInfo: nil,}
	}

	scanner := testBlockScanner(symbol)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	result, contractResult, err := scanner.ExtractTransactionAndReceiptData(txid, scanTargetFunc)
	if err != nil {
		t.Errorf("ExtractTransactionData unexpected error %v", err)
		return
	}

	for sourceKey, keyData := range result {
		log.Notice("account:", sourceKey)
		for _, data := range keyData {

			for i, input := range data.TxInputs {
				log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
			}

			for i, output := range data.TxOutputs {
				log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
			}

			log.Std.Notice("data.Transaction: %+v", data.Transaction)
		}
	}

	for sourceKey, keyData := range contractResult {
		log.Notice("sourceKey:", sourceKey)
		log.Std.Notice("data.ContractTransaction: %+v", keyData)

		for i, event := range keyData.Events {
			log.Std.Notice("data.Contract[%d]: %+v", i, event.Contract)
			log.Std.Notice("data.Events[%d]: %+v", i, event)
		}
	}
}

func TestSubscribeAddress_Contract(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "XUPER"
		addrs      = make(map[string]openwallet.ScanTargetResult)
	)

	contract := &openwallet.SmartContract{
		Symbol:     "XUPER",
		Address:    "0x550cdb1020046b3115a4f8ccebddfb28b66beb27",
		Decimals:   2,
		ContractID: "dl8WD7bM7xk4ZxRybuHCo3JDDtZn2ugPusapoKnQEWA=",
	}
	contract.SetABI(`[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"AddMerchant","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"merchant","type":"address"},{"indexed":true,"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"indexed":true,"internalType":"bytes32","name":"number","type":"bytes32"},{"indexed":false,"internalType":"bytes32","name":"productID","type":"bytes32"}],"name":"AddPrize","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes32","name":"number","type":"bytes32"},{"indexed":false,"internalType":"uint256","name":"price","type":"uint256"},{"indexed":false,"internalType":"address","name":"seller","type":"address"}],"name":"AuctionPrize","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"bytes32","name":"orderNum","type":"bytes32"}],"name":"Burn","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes32","name":"number","type":"bytes32"},{"indexed":false,"internalType":"address","name":"seller","type":"address"},{"indexed":false,"internalType":"address","name":"buyer","type":"address"},{"indexed":false,"internalType":"uint256","name":"dealPrice","type":"uint256"}],"name":"DealAuction","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":true,"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"indexed":true,"internalType":"bytes32","name":"number","type":"bytes32"},{"indexed":false,"internalType":"bytes32","name":"productID","type":"bytes32"},{"indexed":false,"internalType":"uint256","name":"index","type":"uint256"},{"indexed":true,"internalType":"address","name":"winner","type":"address"}],"name":"DrawLotteryPoolPrize","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"address","name":"contractAddress","type":"address"}],"name":"InitLotteryPoolManager","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"address","name":"contractAddress","type":"address"}],"name":"InitWinPrizeManager","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"bytes32","name":"orderNum","type":"bytes32"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"}],"name":"NewLotteryPool","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes32","name":"number","type":"bytes32"},{"indexed":false,"internalType":"address","name":"winner","type":"address"}],"name":"ReceivePrize","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"RemoveMerchant","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"merchant","type":"address"},{"indexed":true,"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"indexed":true,"internalType":"bytes32","name":"number","type":"bytes32"}],"name":"RemovePrize","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"merchant","type":"address"},{"indexed":true,"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"indexed":false,"internalType":"uint8","name":"status","type":"uint8"},{"indexed":false,"internalType":"uint256","name":"drawPrice","type":"uint256"}],"name":"SetLotteryPoolInfo","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"addMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes32","name":"orderNum","type":"bytes32"}],"name":"burn","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"balanceHolder","type":"address"}],"name":"getBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"owner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"isMerchant","outputs":[{"internalType":"bool","name":"flag","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes32","name":"orderNum","type":"bytes32"}],"name":"issue","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"removeMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"supply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"value","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"initManager","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getLotteryPoolManager","outputs":[{"internalType":"address","name":"managerAddress","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"getWinPrizeManager","outputs":[{"internalType":"address","name":"managerAddress","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"}],"name":"newLotteryPool","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"internalType":"uint8","name":"status","type":"uint8"},{"internalType":"uint256","name":"drawPrice","type":"uint256"}],"name":"setLotteryPoolInfo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"internalType":"bytes32","name":"number","type":"bytes32"},{"internalType":"bytes32","name":"productID","type":"bytes32"}],"name":"addPrize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"internalType":"bytes32","name":"number","type":"bytes32"}],"name":"removePrize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"},{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"}],"name":"drawLotteryPoolPrize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes32","name":"number","type":"bytes32"},{"internalType":"uint256","name":"price","type":"uint256"}],"name":"auctionPrize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes32","name":"number","type":"bytes32"},{"internalType":"uint256","name":"dealPrice","type":"uint256"}],"name":"dealAuction","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes32","name":"number","type":"bytes32"}],"name":"receivePrize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"},{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"}],"name":"getLotteryPoolInfo","outputs":[{"internalType":"uint8","name":"status","type":"uint8"},{"internalType":"uint256","name":"prizeSize","type":"uint256"},{"internalType":"uint256","name":"drawPrice","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"},{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getLotteryPoolPrizeByIndex","outputs":[{"internalType":"bytes32","name":"number","type":"bytes32"},{"internalType":"bytes32","name":"productID","type":"bytes32"},{"internalType":"uint256","name":"prizeIndex","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"},{"internalType":"bytes32","name":"lotteryPoolID","type":"bytes32"},{"internalType":"bytes32","name":"num","type":"bytes32"}],"name":"getLotteryPoolPrizeByNumber","outputs":[{"internalType":"bytes32","name":"number","type":"bytes32"},{"internalType":"bytes32","name":"productID","type":"bytes32"},{"internalType":"uint256","name":"prizeIndex","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"bytes32","name":"num","type":"bytes32"}],"name":"getWinPrizeInfo","outputs":[{"internalType":"bytes32","name":"number","type":"bytes32"},{"internalType":"bytes32","name":"productID","type":"bytes32"},{"internalType":"uint8","name":"status","type":"uint8"},{"internalType":"address","name":"winner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"bytes32","name":"num","type":"bytes32"}],"name":"getAuctionInfo","outputs":[{"internalType":"bytes32","name":"number","type":"bytes32"},{"internalType":"uint256","name":"price","type":"uint256"},{"internalType":"uint256","name":"dealPrice","type":"uint256"},{"internalType":"address","name":"buyer","type":"address"},{"internalType":"address","name":"seller","type":"address"},{"internalType":"uint8","name":"status","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	addrs[contract.Address] = openwallet.ScanTargetResult{SourceKey: contract.ContractID, Exist: true, TargetInfo: contract}
	addrs["0xe6a9cc4fe66e7b726e3e8ef8e32c308ce74c0996"] = openwallet.ScanTargetResult{SourceKey: "5kTzFGuAH9UkB9yhZdmXtF8hGPh6iPt4hf8Q3DVy3Xo3", Exist: true,}
	scanTargetFunc := func(target openwallet.ScanTargetParam) openwallet.ScanTargetResult {
		return addrs[target.ScanTarget]
	}

	scanner := testBlockScanner(symbol)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanTargetFuncV2(scanTargetFunc)
	scanner.SetRescanBlockHeight(25249)
	scanner.Run()

	<-endRunning
}

func testBlockScanner(symbol string) openwallet.BlockScanner {
	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return nil
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return nil
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	if scanner.SupportBlockchainDAI() {
		dbFilePath := filepath.Join("data", "db")
		dbFileName := "blockchain.db"
		file.MkdirAll(dbFilePath)
		dai, err := openwallet.NewBlockchainLocal(filepath.Join(dbFilePath, dbFileName), false)
		if err != nil {
			log.Error("NewBlockchainLocal err: %v", err)
			return nil
		}

		scanner.SetBlockchainDAI(dai)
	}
	sub := subscriberSingle{}
	scanner.AddObserver(&sub)

	return scanner
}
