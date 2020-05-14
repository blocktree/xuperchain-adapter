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
	"encoding/base64"
	"encoding/hex"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/pb"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestCallSmartContractABI(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WLEckXM65xfK4BgBKHRfiG4yheRMoax4EX"
	//accountID := "9vEjJLbcrP1bg1MJRTAbPYe4ZwDxamawhPCVEXPi3CMq"
	accountID := "FAyzXxWhEfQ6rWJpNL6agYd9EBvgsKcKgzEWGPgjDbkF"

	contract := openwallet.SmartContract{
		Address: "wasm:counter2",
		Symbol:  "XUPER",
	}
	contract.SetABI(`[{"constant":false,"inputs":[{"name":"key","type":"string"}],"name":"increase","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"key","type":"string"}],"name":"get","outputs":[{"name":"value","type":"string"}],"payable":false,"type":"function"}]`)
	callParam := []string{
		"get",
		"example",
	}

	result, err := tm.CallSmartContractABI(testApp, walletID, accountID, &contract, callParam)
	if err != nil {
		t.Errorf("CallSmartContractABI failed, unexpected error: %v", err)
		return
	}
	log.Infof("result: %+v", result)
	//0x19a4b5d6ea319a5d5ad1d4cc00a5e2e28cac5ec3
}

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
	contractName := "counter7"
	contractCode, ioErr := ioutil.ReadFile(filepath.Join("openw_data", "wasm", "counter"))
	if ioErr != nil {
		t.Errorf("get wasm contract code error: %v", ioErr)
		return
	}
	desc := &pb.WasmCodeDesc{
		Runtime: "go",
	}
	contractDesc, _ := proto.Marshal(desc)

	initarg := `{"creator":"` + base64.StdEncoding.EncodeToString([]byte("xchain")) + `"}`

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



func TestInvokeWasmContract(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "WLEckXM65xfK4BgBKHRfiG4yheRMoax4EX"
	//accountID := "9vEjJLbcrP1bg1MJRTAbPYe4ZwDxamawhPCVEXPi3CMq"
	accountID := "FAyzXxWhEfQ6rWJpNL6agYd9EBvgsKcKgzEWGPgjDbkF"

	contract := openwallet.SmartContract{
		Address: "wasm:counter7",
		Symbol:  "XUPER",
	}
	contract.SetABI(`[{"constant":false,"inputs":[{"name":"key","type":"string"}],"name":"increase","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[{"name":"key","type":"string"}],"name":"get","outputs":[{"name":"value","type":"string"}],"payable":false,"type":"function"}]`)
	callParam := []string{
		"increase",
		"example",
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