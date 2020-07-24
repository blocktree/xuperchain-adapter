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

package xuperchain_rpc

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/xuperchain/xuperchain/core/pb"
	"math/big"
	"strings"
	"testing"
)

var (
	tc = initTestClient()
)

func initTestClient() *Client {
	url := "127.0.0.1:37101"
	chainName := "xuper"
	return NewClient(url, chainName)
}

func TestClient_GetBalanceDetail(t *testing.T) {
	address := "UbFfJuN4U6SqLcVGmJ2kUmgj59sHAd1a5"
	balances, err := tc.GetBalanceDetail(address)
	if err != nil {
		t.Errorf("GetBalanceDetail failed, err: %v", err)
		return
	}
	for _, b := range balances {
		log.Infof("balance: %+v", b)
	}
}

func TestClient_GetBalance(t *testing.T) {
	address := "ahsTENdPBruBtjjJF53ioHAx1yk2HhjnU"
	//address := "XC3333333333333333@xuper"
	balances, err := tc.GetBalance(address)
	if err != nil {
		t.Errorf("GetBalance failed, err: %v", err)
		return
	}
	log.Infof("balance: %+v", balances)
}

func TestClient_GetBlockByHeight(t *testing.T) {
	height := 258917
	block, err := tc.GetBlockByHeight(int64(height))
	if err != nil {
		t.Errorf("GetBlockByHeight failed, err: %v", err)
		return
	}
	txtJson, _ := json.Marshal(block)
	log.Infof("block: %v", string(txtJson))
}

func TestClient_GetBlock(t *testing.T) {
	hash := "667be24256be3fbc7adb7e38d0ba803a714a61e309b2d15ba673112713333a0f"
	block, err := tc.GetBlock(hash)
	if err != nil {
		t.Errorf("GetBlock failed, err: %v", err)
		return
	}
	log.Infof("block: %+v", block)
}

func TestClient_GetBlockChainStatus(t *testing.T) {
	status, err := tc.GetBlockChainStatus()
	if err != nil {
		t.Errorf("GetBlockChainStatus failed, err: %v", err)
		return
	}
	log.Infof("GetMeta: %+v", status.GetMeta())
	log.Infof("GetBlock: %+v", status.GetBlock())
	log.Infof("GetUtxoMeta: %+v", status.GetUtxoMeta())
	log.Infof("BlockID: %+v", hex.EncodeToString(status.GetBlock().GetBlockid()))
}

func TestClient_GetBlockChains(t *testing.T) {
	chains, err := tc.GetBlockChains()
	if err != nil {
		t.Errorf("GetBlockChains failed, err: %v", err)
		return
	}
	log.Infof("chains: %+v", chains)
}

func TestClient_QueryTx(t *testing.T) {
	txid := "8b857276f4b629d0bff671b35268bce7199a11c68dd96f3fa5e7cbef5e5a95d3"
	tx, err := tc.QueryTx(txid)
	if err != nil {
		t.Errorf("QueryTx failed, err: %v", err)
		return
	}
	txjson, _ := json.Marshal(tx.Tx)
	fmt.Printf("\n %s \n", string(txjson))

	var nTx pb.Transaction
	err = json.Unmarshal(txjson, &nTx)
	if err != nil {
		t.Errorf("json.Unmarshal failed, err: %v", err)
		return
	}
	fmt.Printf("blockhash: %s \n", hex.EncodeToString(tx.Tx.Blockid))
	if nTx.String() != tx.Tx.String() {
		fmt.Printf("\n %s \n", tx.Tx.String())
		t.Errorf("json.Unmarshal tx is not equal to original")
		return
	}

}

func TestClient_QueryACL(t *testing.T) {
	account := "XC2222222222222222@xuper"
	acl, isExist, err := tc.QueryACL(account)
	fmt.Printf("Exist: %v", isExist)
	if err != nil {
		t.Errorf("QueryACL failed, err: %v", err)
		return
	}
	objJson, _ := json.Marshal(acl)
	fmt.Printf("\n %s \n", string(objJson))
}

func TestClient_SelectUTXO(t *testing.T) {
	address := "UbFfJuN4U6SqLcVGmJ2kUmgj59sHAd1a5"
	utxo, err := tc.SelectUTXO(address, "10000000", false)
	if err != nil {
		t.Errorf("SelectUTXO failed, err: %v", err)
		return
	}
	for _, u := range utxo {
		log.Infof("utxo: %+v", u)
	}
}

func TestClient_SelectUTXOBySize(t *testing.T) {
	address := "UGbV2vBqMFH4teW7GeaEz19nt7pA5CuT3"
	utxo, err := tc.SelectUTXOBySize(address, false)
	if err != nil {
		t.Errorf("SelectUTXOBySize failed, err: %v", err)
		return
	}
	for _, u := range utxo {
		amount := new(big.Int)
		amount.SetBytes(u.Amount)
		log.Infof("utxo.amount: %s", amount.String())
		log.Infof("utxo.addr: %s", string(u.ToAddr))
		log.Infof("utxo.pubkey: %s", hex.EncodeToString(u.ToPubkey))
	}
}

func TestSplitString(t *testing.T) {
	str := "xkernel"
	args := strings.Split(str, ":")
	log.Infof("args: %+v", args)

	str2 := "wasm:hello"
	args2 := strings.Split(str2, ":")
	log.Infof("args2: %+v", args2)
}

func TestDecodeAddr(t *testing.T) {
	addr := "VUdiVjJ2QnFNRkg0dGVXN0dlYUV6MTludDdwQTVDdVQz"
	addrBit, _ := base64.StdEncoding.DecodeString(addr)
	log.Infof("addr: %s", string(addrBit))
}