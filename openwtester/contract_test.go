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

package openwtester

import (
	"encoding/hex"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/pb"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WLEckXM65xfK4BgBKHRfiG4yheRMoax4EX"
	//accountID := "9vEjJLbcrP1bg1MJRTAbPYe4ZwDxamawhPCVEXPi3CMq"
	accountID := "FAyzXxWhEfQ6rWJpNL6agYd9EBvgsKcKgzEWGPgjDbkF"
	contract := openwallet.SmartContract{
		Address: "xkernel",
		Symbol:  "XUPER",
	}
	contract.SetABI(`[{"constant":false,"inputs":[{"name":"account_name","type":"string"},{"name":"acl","type":"string"}],"name":"NewAccount","outputs":[],"payable":false,"type":"function"}]`)
	accountName := "4444444444444444"
	defaultACL := `
        {
            "pm": {
                "rule": 1,
                "acceptValue": 1.0
            },
            "aksWeight": {
                "ahsTENdPBruBtjjJF53ioHAx1yk2HhjnU": 1.0
            }
        }
        `

	callParam := []string{
		"NewAccount",
		accountName,
		defaultACL,
	}

	rawTx, err := tm.CreateSmartContractTransaction(testApp, walletID, accountID, "", "", &contract, callParam)
	if err != nil {
		t.Errorf("CreateSmartContractTransaction failed, unexpected error: %v", err)
		return
	}
	//log.Infof("rawTx: %+v", rawTx)

	_, err = tm.SignSmartContractTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		t.Errorf("SignSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	tx, err := tm.SubmitSmartContractTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		t.Errorf("SubmitSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)
}

func TestDeployWasmContract(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WLEckXM65xfK4BgBKHRfiG4yheRMoax4EX"
	//accountID := "9vEjJLbcrP1bg1MJRTAbPYe4ZwDxamawhPCVEXPi3CMq"
	accountID := "FAyzXxWhEfQ6rWJpNL6agYd9EBvgsKcKgzEWGPgjDbkF"
	contract := openwallet.SmartContract{
		Address: "xkernel",
		Symbol:  "XUPER",
	}
	contract.SetABI(`[{"constant":false,"inputs":[{"name":"account_name","type":"string"},{"name":"contract_name","type":"string"},{"name":"contract_code","type":"bytes"},{"name":"contract_desc","type":"bytes"},{"name":"init_args","type":"string"}],"name":"Deploy","outputs":[],"payable":false,"type":"function"}]`)
	accountName := "XC3333333333333333@xuper"
	contractName := "artToyContract2"
	contractCode, ioErr := ioutil.ReadFile(filepath.Join("openw_data", "wasm", "artToyContract.wasm"))
	if ioErr != nil {
		t.Errorf("get wasm contract code error: %v", ioErr)
		return
	}
	desc := &pb.WasmCodeDesc{
		Runtime: "go",
	}
	contractDesc, _ := proto.Marshal(desc)

	//initarg := `{"initSupply":"` + base64.StdEncoding.EncodeToString([]byte("20000000000")) + `"}`
	initarg := `{}`

	callParam := []string{
		"Deploy",
		accountName,
		contractName,
		hex.EncodeToString(contractCode),
		hex.EncodeToString(contractDesc),
		initarg,
	}

	rawTx, err := tm.CreateSmartContractTransaction(testApp, walletID, accountID, "", "", &contract, callParam)
	if err != nil {
		t.Errorf("CreateSmartContractTransaction failed, unexpected error: %v", err)
		return
	}
	//log.Infof("rawTx: %+v", rawTx)

	_, err = tm.SignSmartContractTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		t.Errorf("SignSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	tx, err := tm.SubmitSmartContractTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		t.Errorf("SubmitSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)
}



func TestContractIssue(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WLEckXM65xfK4BgBKHRfiG4yheRMoax4EX"
	//accountID := "9vEjJLbcrP1bg1MJRTAbPYe4ZwDxamawhPCVEXPi3CMq"
	accountID := "FAyzXxWhEfQ6rWJpNL6agYd9EBvgsKcKgzEWGPgjDbkF"

	contract := openwallet.SmartContract{
		Address: "wasm:artToyContract2",
		Symbol:  "XUPER",
	}
	contract.SetABI(`[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"from","type":"address"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"string","name":"orderNum","type":"string"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"string","name":"orderNum","type":"string"}],"name":"Burn","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"orderNum","type":"string"}],"name":"issue","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"orderNum","type":"string"}],"name":"burn","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"address","name":"owner","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"allowance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"address","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"totalSupply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":false,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"owner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"AddMerchant","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"RemoveMerchant","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"addMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"removeMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"isMerchant","outputs":[{"internalType":"bool","name":"flag","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"AddArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"}],"name":"RemoveArtToyFromSeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"uint64","name":"index","type":"uint64"},{"indexed":false,"internalType":"string","name":"owner","type":"string"}],"name":"PurchaseArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"string","name":"owner","type":"string"}],"name":"ReceiveArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"ReturnArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"uint64","name":"status","type":"uint64"},{"indexed":false,"internalType":"uint256","name":"drawPrice","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"returnPrice","type":"uint256"}],"name":"SetArtToySeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"RevokeArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"}],"name":"RevokeArtToySeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"orderNumber","type":"string"},{"indexed":false,"internalType":"string","name":"sellerAddr","type":"string"},{"indexed":false,"internalType":"string","name":"buyerAddr","type":"string"},{"indexed":false,"internalType":"uint256","name":"offerAmount","type":"uint256"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"uint64","name":"orderType","type":"uint64"},{"indexed":false,"internalType":"uint64","name":"status","type":"uint64"}],"name":"OrderArtToyExchange","type":"event"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"isArtToySeriesExist","outputs":[{"internalType":"bool","name":"flag","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"}],"name":"newArtToySeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"}],"name":"addArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"}],"name":"removeArtToyFromSeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"}],"name":"revokeArtToySeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"revokeArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"uint256","name":"drawPrice","type":"uint256"},{"internalType":"uint256","name":"returnPrice","type":"uint256"},{"internalType":"uint256","name":"status","type":"uint256"}],"name":"setArtToySeriesInfo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"merchant","type":"string"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"purchaseArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"receiveArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"returnArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"orderType","type":"uint256"}],"name":"orderArtToyExchange","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"}],"name":"getArtToyExchangeOrder","outputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"sellerAddr","type":"string"},{"internalType":"string","name":"buyerAddr","type":"string"},{"internalType":"uint256","name":"offerAmount","type":"uint256"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint64","name":"orderType","type":"uint64"},{"internalType":"uint64","name":"status","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"number","type":"string"}],"name":"dealArtToyExchange","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"}],"name":"cancelArtToyExchangeOrder","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"getArtToyByNumber","outputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint256","name":"merchant","type":"uint256"},{"internalType":"string","name":"owner","type":"string"},{"internalType":"uint64","name":"creatAt","type":"uint64"},{"internalType":"uint64","name":"status","type":"uint64"},{"internalType":"uint64","name":"index","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"merchant","type":"string"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"getArtToySeries","outputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"uint256","name":"merchant","type":"uint256"},{"internalType":"uint64","name":"status","type":"uint64"},{"internalType":"uint256","name":"drawPrice","type":"uint256"},{"internalType":"uint256","name":"returnPrice","type":"uint256"},{"internalType":"uint64","name":"toySize","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	callParam := []string{
		"issue",
		"Rvm1AE6rZwLpPFbcBfD7wZxXK3FR6QEXb",
		"999",
		"order1234",
	}

	rawTx, err := tm.CreateSmartContractTransaction(testApp, walletID, accountID, "", "", &contract, callParam)
	if err != nil {
		t.Errorf("CreateSmartContractTransaction failed, unexpected error: %v", err)
		return
	}
	//log.Infof("rawTx: %+v", rawTx)

	_, err = tm.SignSmartContractTransaction(testApp, walletID, accountID, "12345678", rawTx)
	if err != nil {
		t.Errorf("SignSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	rawTx.AwaitResult = true
	rawTx.AwaitTimeout = 10
	tx, err := tm.SubmitSmartContractTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		t.Errorf("SubmitSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("txID:", rawTx.TxID)

	for i, event := range tx.Events {
		log.Std.Notice("data.Events[%d]: %+v", i, event)
	}
}


func TestInvokeWasmContractDouble(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WLEckXM65xfK4BgBKHRfiG4yheRMoax4EX"
	accountID1 := "9vEjJLbcrP1bg1MJRTAbPYe4ZwDxamawhPCVEXPi3CMq"
	accountID2 := "FAyzXxWhEfQ6rWJpNL6agYd9EBvgsKcKgzEWGPgjDbkF"

	contract := openwallet.SmartContract{
		Address: "wasm:mytoken11",
		Symbol:  "XUPER",
	}
	contract.SetABI(`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"bytes32","name":"orderNum","type":"bytes32"}],"name":"Burn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[],"name":"Pause","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"string","name":"from","type":"address"},{"indexed":true,"internalType":"string","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"string","name":"player","type":"address"},{"indexed":true,"internalType":"string","name":"game","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Game","type":"event"},{"anonymous":false,"inputs":[],"name":"Unpause","type":"event"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"owner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"pause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[],"name":"unpause","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"issue","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes32","name":"orderNum","type":"bytes32"}],"name":"burn","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"game","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"owner","type":"string"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"supply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"balanceHolder","type":"address"}],"name":"getBalance","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	callParam := []string{
		"game",
		"WayY2YjxdvMFwj6PubSh3waZEWZamz1tJ",
		"123",
	}

	rawTx, err := tm.CreateSmartContractTransaction(testApp, walletID, accountID1, "", "", &contract, callParam)
	if err != nil {
		t.Errorf("CreateSmartContractTransaction failed, unexpected error: %v", err)
		return
	}
	//log.Infof("rawTx: %+v", rawTx)

	_, err = tm.SignSmartContractTransaction(testApp, walletID, rawTx.Account.AccountID, "12345678", rawTx)
	if err != nil {
		t.Errorf("SignSmartContractTransaction failed, unexpected error: %v", err)
		return
	}


	rawTx2, err := tm.CreateSmartContractTransaction(testApp, walletID, accountID2, "", "", &contract, callParam)
	if err != nil {
		t.Errorf("CreateSmartContractTransaction failed, unexpected error: %v", err)
		return
	}
	//log.Infof("rawTx: %+v", rawTx)

	_, err = tm.SignSmartContractTransaction(testApp, walletID, rawTx.Account.AccountID, "12345678", rawTx2)
	if err != nil {
		t.Errorf("SignSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	tx, err := tm.SubmitSmartContractTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		t.Errorf("SubmitSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	log.Info("txID:", tx.TxID)

	tx2, err := tm.SubmitSmartContractTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx2)
	if err != nil {
		t.Errorf("SubmitSmartContractTransaction failed, unexpected error: %v", err)
		return
	}

	log.Info("txID2:", tx2.TxID)
}


func TestContractBalanceOf(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WLEckXM65xfK4BgBKHRfiG4yheRMoax4EX"
	//accountID := "9vEjJLbcrP1bg1MJRTAbPYe4ZwDxamawhPCVEXPi3CMq"
	accountID := "FAyzXxWhEfQ6rWJpNL6agYd9EBvgsKcKgzEWGPgjDbkF"

	contract := openwallet.SmartContract{
		Address: "wasm:artToyContract2",
		Symbol:  "XUPER",
	}
	contract.SetABI(`[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"from","type":"address"},{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"string","name":"orderNum","type":"string"}],"name":"Issue","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"string","name":"orderNum","type":"string"}],"name":"Burn","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"orderNum","type":"string"}],"name":"issue","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"string","name":"orderNum","type":"string"}],"name":"burn","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"address","name":"owner","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"allowance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"address","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"totalSupply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":false,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getOwner","outputs":[{"internalType":"address","name":"owner","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"AddMerchant","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"}],"name":"RemoveMerchant","type":"event"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"addMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"removeMerchant","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"}],"name":"isMerchant","outputs":[{"internalType":"bool","name":"flag","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"AddArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"}],"name":"RemoveArtToyFromSeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"merchant","type":"address"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"uint64","name":"index","type":"uint64"},{"indexed":false,"internalType":"string","name":"owner","type":"string"}],"name":"PurchaseArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"string","name":"owner","type":"string"}],"name":"ReceiveArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"ReturnArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"},{"indexed":false,"internalType":"uint64","name":"status","type":"uint64"},{"indexed":false,"internalType":"uint256","name":"drawPrice","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"returnPrice","type":"uint256"}],"name":"SetArtToySeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"}],"name":"RevokeArtToy","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"merchant","type":"string"},{"indexed":false,"internalType":"string","name":"seriesID","type":"string"}],"name":"RevokeArtToySeries","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"orderNumber","type":"string"},{"indexed":false,"internalType":"string","name":"sellerAddr","type":"string"},{"indexed":false,"internalType":"string","name":"buyerAddr","type":"string"},{"indexed":false,"internalType":"uint256","name":"offerAmount","type":"uint256"},{"indexed":false,"internalType":"string","name":"number","type":"string"},{"indexed":false,"internalType":"string","name":"productID","type":"string"},{"indexed":false,"internalType":"uint64","name":"orderType","type":"uint64"},{"indexed":false,"internalType":"uint64","name":"status","type":"uint64"}],"name":"OrderArtToyExchange","type":"event"},{"constant":true,"inputs":[{"internalType":"address","name":"merchant","type":"address"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"isArtToySeriesExist","outputs":[{"internalType":"bool","name":"flag","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"}],"name":"newArtToySeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"}],"name":"addArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"}],"name":"removeArtToyFromSeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"}],"name":"revokeArtToySeries","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"revokeArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"uint256","name":"drawPrice","type":"uint256"},{"internalType":"uint256","name":"returnPrice","type":"uint256"},{"internalType":"uint256","name":"status","type":"uint256"}],"name":"setArtToySeriesInfo","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"merchant","type":"string"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"purchaseArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"receiveArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"returnArtToy","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"orderType","type":"uint256"}],"name":"orderArtToyExchange","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"}],"name":"getArtToyExchangeOrder","outputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"sellerAddr","type":"string"},{"internalType":"string","name":"buyerAddr","type":"string"},{"internalType":"uint256","name":"offerAmount","type":"uint256"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint64","name":"orderType","type":"uint64"},{"internalType":"uint64","name":"status","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"},{"internalType":"string","name":"number","type":"string"}],"name":"dealArtToyExchange","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"string","name":"orderNumber","type":"string"}],"name":"cancelArtToyExchangeOrder","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"number","type":"string"}],"name":"getArtToyByNumber","outputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"string","name":"number","type":"string"},{"internalType":"string","name":"productID","type":"string"},{"internalType":"uint256","name":"merchant","type":"uint256"},{"internalType":"string","name":"owner","type":"string"},{"internalType":"uint64","name":"creatAt","type":"uint64"},{"internalType":"uint64","name":"status","type":"uint64"},{"internalType":"uint64","name":"index","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"string","name":"merchant","type":"string"},{"internalType":"string","name":"seriesID","type":"string"}],"name":"getArtToySeries","outputs":[{"internalType":"string","name":"seriesID","type":"string"},{"internalType":"uint256","name":"merchant","type":"uint256"},{"internalType":"uint64","name":"status","type":"uint64"},{"internalType":"uint256","name":"drawPrice","type":"uint256"},{"internalType":"uint256","name":"returnPrice","type":"uint256"},{"internalType":"uint64","name":"toySize","type":"uint64"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	callParam := []string{
		"balanceOf",
		"Rvm1AE6rZwLpPFbcBfD7wZxXK3FR6QEXb",
	}

	result, err := tm.CallSmartContractABI(testApp, walletID, accountID, &contract, callParam)
	if err != nil {
		t.Errorf("CallSmartContractABI failed, unexpected error: %v", err)
		return
	}
	log.Infof("result: %+v", result)

}