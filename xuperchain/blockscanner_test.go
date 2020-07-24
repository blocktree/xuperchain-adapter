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
	"encoding/hex"
	"github.com/blocktree/openwallet/log"
	"testing"
)

func TestBlockScanner_tx_outputs_ext(t *testing.T) {
	//txid := "ec3044f3a5aaa6a8472101babb6bb575327b96e2ae32e6cd2181ca0a159d2485"
	txid := "f123cb6b8b3a8601dc798a2ab6d76a0abf270091bdc1717914bcb4d3f27be304"
	tx, err := tw.RPC.QueryTx(txid)
	if err != nil {
		t.Errorf("QueryTx failed, err: %v", err)
		return
	}

	txInputExts := tx.Tx.GetTxInputsExt()
	for _, ext := range txInputExts {
		log.Infof("input bucket: %s", ext.GetBucket())
		log.Infof("input key: %s", string(ext.GetKey()))
		log.Infof("input reftxid: %s", hex.EncodeToString(ext.GetRefTxid()))
		log.Infof("input refoffset: %v", ext.GetRefOffset())
		log.Infof("input version: %v", ext.GetVersion())
	}

	txOutExts := tx.Tx.GetTxOutputsExt()
	for _, ext := range txOutExts {
		log.Infof("output bucket: %s", ext.GetBucket())
		log.Infof("output key: %s", string(ext.GetKey()))
		log.Infof("output value: %s", string(ext.GetValue()))
	}

}
