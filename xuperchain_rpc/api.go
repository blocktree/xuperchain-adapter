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
package xuperchain_rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/xuperchain/xuper-sdk-go/common"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
	"google.golang.org/grpc"
	"time"
)

type Client struct {
	BaseURL      string
	xchainClient pb.XchainClient
	ChainName    string
	timeout      time.Duration
}

func NewClient(url, chainName string) *Client {

	client := &Client{
		BaseURL:   url,
		ChainName: chainName,
	}

	return client
}

func (xc *Client) connect() error {

	if xc.xchainClient != nil {
		return nil
	}

	if len(xc.BaseURL) == 0 {
		return fmt.Errorf("BaseURL is empty")
	}

	if len(xc.ChainName) == 0 {
		return fmt.Errorf("ChainName is empty")
	}

	conn, err := grpc.Dial(xc.BaseURL, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	xchainClient := pb.NewXchainClient(conn)
	xc.xchainClient = xchainClient
	xc.timeout = 60 * time.Second

	return nil
}

// GetBalanceDetail
func (xc *Client) GetBalanceDetail(address string) ([]*pb.TokenFrozenDetail, error) {

	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	tfds := []*pb.TokenFrozenDetails{{Bcname: xc.ChainName}}
	addStatus := &pb.AddressBalanceStatus{
		Address: address,
		Tfds:    tfds,
	}

	res, err := xc.xchainClient.GetBalanceDetail(ctx, addStatus)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}

	for _, bc := range res.GetTfds() {
		if bc.Bcname == xc.ChainName {
			return bc.Tfd, nil
		}
	}

	return nil, nil
}

// GetBalance
func (xc *Client) GetBalance(address string) (*pb.TokenDetail, error) {

	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	bc := &pb.TokenDetail{
		Bcname: xc.ChainName,
	}

	in := &pb.AddressStatus{
		Address: address,
		Bcs:     []*pb.TokenDetail{bc},
	}

	res, err := xc.xchainClient.GetBalance(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}

	for _, bc := range res.GetBcs() {
		if bc.Bcname == xc.ChainName {
			return bc, nil
		}
	}

	return nil, nil

}

func (xc *Client) GetBlock(hash string) (*pb.InternalBlock, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	id, _ := hex.DecodeString(hash)
	in := &pb.BlockID{
		Bcname:      xc.ChainName,
		Blockid:     id,
		NeedContent: true,
	}

	res, err := xc.xchainClient.GetBlock(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}
	return res.GetBlock(), nil
}

func (xc *Client) GetBlockByHeight(height int64) (*pb.InternalBlock, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	in := &pb.BlockHeight{
		Bcname: xc.ChainName,
		Height: height,
	}

	res, err := xc.xchainClient.GetBlockByHeight(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}
	return res.GetBlock(), nil
}

func (xc *Client) GetBlockChainStatus() (*pb.BCStatus, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	in := &pb.BCStatus{
		Bcname: xc.ChainName,
	}
	res, err := xc.xchainClient.GetBlockChainStatus(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}
	return res, nil
}

func (xc *Client) GetBlockChains() ([]string, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	in := &pb.CommonIn{}
	res, err := xc.xchainClient.GetBlockChains(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}
	return res.GetBlockchains(), nil
}

func (xc *Client) GetSystemStatus() ([]*pb.BCStatus, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	in := &pb.CommonIn{}
	res, err := xc.xchainClient.GetSystemStatus(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}
	return res.GetSystemsStatus().GetBcsStatus(), nil
}

func (xc *Client) QueryTx(txid string) (*pb.TxStatus, error) {

	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	id, _ := hex.DecodeString(txid)
	in := &pb.TxStatus{
		Bcname: xc.ChainName,
		Txid:   id,
	}
	res, err := xc.xchainClient.QueryTx(ctx, in)
	if err != nil {
		return nil, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}
	if res.Tx == nil {
		return nil, common.ErrTxNotFound
	}
	return res, nil
}

func (xc *Client) QueryACL(accountName string) (*pb.AclStatus, bool, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, false, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	in := &pb.AclStatus{
		Bcname:      xc.ChainName,
		AccountName: accountName,
	}
	res, err := xc.xchainClient.QueryACL(ctx, in)
	if err != nil {
		return nil, false, err
	}
	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, false, fmt.Errorf(res.Header.Error.String())
	}

	if !res.Confirmed {
		return nil, false, nil
	}

	return res, true, nil

}

func (xc *Client) PreExec(in *pb.InvokeRPCRequest) (*pb.InvokeRPCResponse, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	res, err := xc.xchainClient.PreExec(ctx, in)
	if err != nil {
		return nil, err
	}

	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}

	for _, res := range res.GetResponse().GetResponses() {
		if res.Status >= 400 {
			return nil, fmt.Errorf("contract error status:%d message:%s", res.Status, res.Message)
		}
	}
	return res, nil

}

func (xc *Client) PreExecWithSelectUTXO(in *pb.PreExecWithSelectUTXORequest) (*pb.PreExecWithSelectUTXOResponse, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	res, err := xc.xchainClient.PreExecWithSelectUTXO(ctx, in)
	if err != nil {
		return nil, err
	}

	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}

	for _, res := range res.GetResponse().GetResponses() {
		if res.Status >= 400 {
			return nil, fmt.Errorf("contract error status:%d message:%s", res.Status, res.Message)
		}
	}
	return res, nil

}

func (xc *Client) SelectUTXO(address, totalNeed string, needLock bool) ([]*pb.Utxo, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	in := &pb.UtxoInput{
		Bcname:    xc.ChainName,
		Address:   address,
		TotalNeed: totalNeed,
		NeedLock:  needLock,
	}

	res, err := xc.xchainClient.SelectUTXO(ctx, in)
	if err != nil {
		return nil, err
	}

	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}

	return res.UtxoList, nil

}

func (xc *Client) SelectUTXOBySize(address string, needLock bool) ([]*pb.Utxo, error) {
	if cErr := xc.connect(); cErr != nil {
		return nil, cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	in := &pb.UtxoInput{
		Bcname:    xc.ChainName,
		Address:   address,
		NeedLock:  needLock,
	}

	res, err := xc.xchainClient.SelectUTXOBySize(ctx, in)
	if err != nil {
		return nil, err
	}

	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, fmt.Errorf(res.Header.Error.String())
	}

	return res.UtxoList, nil

}

func (xc *Client) PostTx(tx *pb.Transaction) (string, error) {
	if cErr := xc.connect(); cErr != nil {
		return "", cErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), xc.timeout)
	defer cancel()

	// 然后和上一节一致了，生成交易ID
	tx.Txid, _ = txhash.MakeTransactionID(tx)

	txStatus := &pb.TxStatus{
		Bcname: xc.ChainName,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Txid:   tx.Txid,
	}

	res, err := xc.xchainClient.PostTx(ctx, txStatus)
	if err != nil {
		return "", err
	}

	if res.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return "", fmt.Errorf(res.Header.Error.String())
	}

	return hex.EncodeToString(tx.Txid), nil

}
