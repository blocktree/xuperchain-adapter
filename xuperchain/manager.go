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
package xuperchain

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/blocktree/xuperchain-adapter/xuperchain_addrdec"
	"github.com/blocktree/xuperchain-adapter/xuperchain_rpc"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/xuperchain/xuperchain/core/pb"
	"math/big"
	"strings"
)

type WalletManager struct {
	openwallet.AssetsAdapterBase

	RPC             *xuperchain_rpc.Client        //节点客户端
	Config          *ChainConfig                  //钱包管理配置
	BlockScanner    openwallet.BlockScanner       //区块扫描器
	AddrDecoder     openwallet.AddressDecoderV2   //地址编码器
	TxDecoder       openwallet.TransactionDecoder //交易单编码器
	ContractDecoder *ContractDecoder              //智能合约解释器
	Log             *log.OWLogger                 //日志工具
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol)
	wm.BlockScanner = NewBlockScanner(&wm)
	wm.AddrDecoder = xuperchain_addrdec.NewAddressDecoder(wm.CurveType())
	wm.TxDecoder = NewTransactionDecoder(&wm)
	wm.ContractDecoder = &ContractDecoder{wm: &wm}
	wm.Log = log.NewOWLogger(wm.Symbol())

	return &wm
}

// EncodeInvokeRequest 编码API调用参数
func (wm *WalletManager) EncodeInvokeRequest(abiInstance abi.ABI, contractAddress string, abiParam ...string) (*pb.InvokeRequest, error) {

	var (
		args         = make(map[string][]byte)
		moduleName   = ""
		contractName = ""
	)

	//拆分合约地址
	contractInfo := strings.Split(contractAddress, ":")

	if len(contractInfo) == 0 {
		return nil, fmt.Errorf("contract address is invalid")
	}

	if len(abiParam) == 0 {
		return nil, fmt.Errorf("abi param length is empty")
	}

	if len(contractInfo) == 2 {
		moduleName = contractInfo[0]
		contractName = contractInfo[1]
	} else {
		moduleName = contractInfo[0]
	}

	method := abiParam[0]
	//转化string参数为abi调用参数
	abiMethod, ok := abiInstance.Methods[method]
	if !ok {
		return nil, fmt.Errorf("abi method can not found")
	}
	abiArgs := abiParam[1:]
	if len(abiMethod.Inputs) != len(abiArgs) {
		return nil, fmt.Errorf("abi input arguments is: %d, except is : %d", len(abiArgs), len(abiMethod.Inputs))
	}
	for i, input := range abiMethod.Inputs {

		var a []byte
		switch input.Type.T {
		case abi.BoolTy:
			if abiArgs[i] == "true" {
				a = []byte{0x01}
			} else {
				a = []byte{0x00}
			}
		case abi.UintTy, abi.IntTy:
			a, _ = convertParamToNum(abiArgs[i])
		case abi.AddressTy:
			a = []byte(abiArgs[i])
		case abi.FixedBytesTy, abi.BytesTy, abi.HashTy:
			a, _ = hex.DecodeString(abiArgs[i])
		case abi.StringTy:
			a = []byte(abiArgs[i])
		}

		args[input.Name] = a
	}

	invokeRequest := &pb.InvokeRequest{
		ModuleName:   moduleName,
		MethodName:   method,
		ContractName: contractName,
		Args:         args,
	}

	return invokeRequest, nil
}


func convertParamToNum(param string) ([]byte, error) {
	var (
		base int
		bInt *big.Int
		err  error
	)
	if strings.HasPrefix(param, "0x") {
		base = 16
	} else {
		base = 10
	}
	bInt, err = common.StringValueToBigInt(param, base)
	if err != nil {
		return nil, err
	}

	return bInt.Bytes(), nil
}