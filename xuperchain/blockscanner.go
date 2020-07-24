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
	"fmt"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"github.com/xuperchain/xuperchain/core/pb"
	"time"
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
	extractData         map[string]*openwallet.TxExtractData //主链交易
	extractContractData map[string]*openwallet.SmartContractReceipt //合约回执
	TxID                string
	BlockHeight         uint64
	Success             bool
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
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//ScanBlockTask 扫描任务
func (bs *BlockScanner) ScanBlockTask() {

	//获取本地区块高度
	blockHeader, err := bs.GetScannedBlockHeader()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block height; unexpected error: %v", err)
		return
	}

	currentHeight := blockHeader.Height
	currentHash := blockHeader.Hash

	for {

		if !bs.Scanning {
			//区块扫描器已暂停，马上结束本次任务
			return
		}

		//获取最大高度
		maxHeight, err := bs.GetBlockHeight()
		if err != nil {
			//下一个高度找不到会报异常
			bs.wm.Log.Std.Info("block scanner can not get rpc-server block height; unexpected error: %v", err)
			break
		}

		//是否已到最新高度
		if currentHeight >= maxHeight {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height: %d", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		bs.wm.Log.Std.Info("block scanner scanning height: %d ...", currentHeight)

		block, err := bs.wm.RPC.GetBlockByHeight(int64(currentHeight))
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

			//记录未扫区块
			unscanRecord := openwallet.NewUnscanRecord(currentHeight, "", err.Error(), bs.wm.Symbol())
			bs.SaveUnscanRecord(unscanRecord)
			bs.wm.Log.Std.Info("block height: %d extract failed.", currentHeight)
			continue
		}

		isFork := false
		preHash := hex.EncodeToString(block.PreHash)
		//判断hash是否上一区块的hash
		if currentHash != preHash {

			bs.wm.Log.Std.Info("block has been fork on height: %d.", currentHeight)
			bs.wm.Log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			bs.wm.Log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, preHash)

			bs.wm.Log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			//查询本地分叉的区块
			forkBlock, _ := bs.GetLocalBlock(currentHeight - 1)

			//删除上一区块链的所有充值记录
			//bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块链的未扫记录
			bs.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.GetLocalBlock(currentHeight)
			if err != nil {
				bs.wm.Log.Std.Error("block scanner can not get local block; unexpected error: %v", err)
				bs.wm.Log.Info("block scanner prev block height:", currentHeight)

				localBlock, err = bs.wm.RPC.GetBlockByHeight(int64(currentHeight))
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}

			}

			//重置当前区块的hash
			currentHash = hex.EncodeToString(localBlock.Blockid)

			bs.wm.Log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalBlockHead(uint64(localBlock.Height), hex.EncodeToString(localBlock.Blockid))

			isFork = true

			if forkBlock != nil {

				//通知分叉区块给观测者，异步处理
				bs.newBlockNotify(forkBlock, isFork)
			}

		} else {

			err = bs.BatchExtractTransaction(block)
			if err != nil {
				bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			}

			//重置当前区块的hash
			currentHash = hex.EncodeToString(block.Blockid)

			//保存本地新高度
			bs.SaveLocalBlockHead(currentHeight, currentHash)
			bs.SaveLocalBlock(block)

			isFork = false

			//通知新区块给观测者，异步处理
			bs.newBlockNotify(block, isFork)
		}

	}

	//重扫前N个块，为保证记录找到
	for i := currentHeight - bs.RescanLastBlockCount; i < currentHeight; i++ {
		bs.scanBlock(i)
	}

	//重扫失败区块
	bs.RescanFailedRecord()

}

//ScanBlock 扫描指定高度区块
func (bs *BlockScanner) ScanBlock(height uint64) error {

	block, err := bs.scanBlock(height)
	if err != nil {
		return err
	}

	//通知新区块给观测者，异步处理
	bs.newBlockNotify(block, false)

	return nil
}

func (bs *BlockScanner) scanBlock(height uint64) (*pb.InternalBlock, error) {

	block, err := bs.wm.RPC.GetBlockByHeight(int64(height))
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)

		//记录未扫区块
		unscanRecord := openwallet.NewUnscanRecord(height, "", err.Error(), bs.wm.Symbol())
		bs.SaveUnscanRecord(unscanRecord)
		bs.wm.Log.Std.Info("block height: %d extract failed.", height)
		return nil, err
	}

	bs.wm.Log.Std.Info("block scanner scanning height: %d ...", block.Height)

	err = bs.BatchExtractTransaction(block)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
	}

	//保存区块
	//bs.wm.SaveLocalBlock(block)

	return block, nil
}

//rescanFailedRecord 重扫失败记录
func (bs *BlockScanner) RescanFailedRecord() {

	var (
		blockMap = make(map[uint64][]string)
	)

	list, err := bs.GetUnscanRecords()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get rescan data; unexpected error: %v", err)
	}

	//组合成批处理
	for _, r := range list {

		if _, exist := blockMap[r.BlockHeight]; !exist {
			blockMap[r.BlockHeight] = make([]string, 0)
		}

		if len(r.TxID) > 0 {
			arr := blockMap[r.BlockHeight]
			arr = append(arr, r.TxID)

			blockMap[r.BlockHeight] = arr
		}
	}

	for height, _ := range blockMap {

		if height == 0 {
			continue
		}

		bs.wm.Log.Std.Info("block scanner rescanning height: %d ...", height)

		block, err := bs.wm.RPC.GetBlockByHeight(int64(height))
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
			continue
		}

		err = bs.BatchExtractTransaction(block)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}

		//删除未扫记录
		bs.DeleteUnscanRecord(height)
	}
}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *BlockScanner) newBlockNotify(block *pb.InternalBlock, isFork bool) {

	obj := &openwallet.BlockHeader{}
	//解析json
	obj.Hash = hex.EncodeToString(block.Blockid)
	obj.Merkleroot = hex.EncodeToString(block.MerkleRoot)
	obj.Previousblockhash = hex.EncodeToString(block.PreHash)
	obj.Height = uint64(block.Height)
	obj.Time = uint64(block.Timestamp)
	obj.Symbol = bs.wm.Symbol()
	obj.Fork = isFork
	bs.NewBlockNotify(obj)
}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (bs *BlockScanner) BatchExtractTransaction(block *pb.InternalBlock) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(block.Transactions) //需要完成的总数
	)

	if len(block.Transactions) == 0 {
		return nil
	}

	//生产通道
	producer := make(chan ExtractResult)
	defer close(producer)

	//消费通道
	worker := make(chan ExtractResult)
	defer close(worker)

	//保存工作
	saveWork := func(height uint64, result chan ExtractResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {

				notifyErr := bs.newExtractDataNotify(height, gets.extractData)
				if notifyErr != nil {
					failed++ //标记保存失败数
					bs.wm.Log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}

			} else {
				//记录未扫区块
				unscanRecord := openwallet.NewUnscanRecord(height, "", "", bs.wm.Symbol())
				bs.SaveUnscanRecord(unscanRecord)
				bs.wm.Log.Std.Info("block height: %d extract failed.", height)
				failed++ //标记保存失败数
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				//bs.wm.Log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(eblockHeight uint64, eBlockHash string, mTxs []*pb.Transaction, eProducer chan ExtractResult) {
		for _, tx := range mTxs {
			bs.extractingCH <- struct{}{}
			//shouldDone++
			go func(mBlockHeight uint64, mTx *pb.Transaction, end chan struct{}, mProducer chan<- ExtractResult) {

				//导出提出的交易
				mProducer <- bs.ExtractTransaction(mBlockHeight, eBlockHash, mTx, bs.ScanTargetFuncV2)
				//释放
				<-end

			}(eblockHeight, tx, bs.extractingCH, eProducer)
		}
	}

	/*	开启导出的线程	*/

	//独立线程运行消费
	go saveWork(uint64(block.Height), worker)

	//独立线程运行生产
	go extractWork(uint64(block.Height), hex.EncodeToString(block.Blockid), block.Transactions, producer)

	//以下使用生产消费模式
	bs.extractRuntime(producer, worker, quit)

	if failed > 0 {
		return fmt.Errorf("block scanner saveWork failed")
	} else {
		return nil
	}

	//return nil
}

//extractRuntime 提取运行时
func (bs *BlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

	var (
		values = make([]ExtractResult, 0)
	)

	for {

		var activeWorker chan<- ExtractResult
		var activeValue ExtractResult

		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			//bs.wm.Log.Std.Info("block scanner have been scanned!")
			return
		case activeWorker <- activeValue:
			//wm.Log.Std.Info("Get %d", len(activeValue))
			values = values[1:]
		}
	}

}

//ExtractTransaction 提取交易单
func (bs *BlockScanner) ExtractTransaction(blockHeight uint64, blockHash string, tx *pb.Transaction, scanAddressFunc openwallet.BlockScanTargetFuncV2) ExtractResult {

	var (
		result = ExtractResult{
			BlockHeight:         blockHeight,
			TxID:                hex.EncodeToString(tx.Txid),
			extractData:         make(map[string]*openwallet.TxExtractData),
			extractContractData: make(map[string]*openwallet.SmartContractReceipt),
		}
	)

	//提取主币交易单
	bs.extractTransaction(blockHeight, tx, &result, scanAddressFunc)
	//提取代币交易单
	bs.extractSmartContractTransaction(blockHeight, tx, &result, scanAddressFunc)
	return result

}

//ExtractTransactionData 提取交易单
func (bs *BlockScanner) extractTransaction(blockHeight uint64, trx *pb.Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanTargetFuncV2) {

	var (
		success = true
	)

	txType := uint64(0)
	txAction := ""

	//提取出账部分记录
	from, totalSpent := bs.extractTxInput(blockHeight, trx, result, scanAddressFunc)
	//bs.wm.Log.Debug("from:", from, "totalSpent:", totalSpent)

	//提取入账部分记录
	to, totalReceived := bs.extractTxOutput(blockHeight, trx, result, scanAddressFunc)
	//bs.wm.Log.Debug("to:", to, "totalReceived:", totalReceived)

	for _, extractData := range result.extractData {
		tx := &openwallet.Transaction{
			From: from,
			To:   to,
			Fees: totalSpent.Sub(totalReceived).StringFixed(8),
			Coin: openwallet.Coin{
				Symbol:     bs.wm.Symbol(),
				IsContract: false,
			},
			BlockHash:   hex.EncodeToString(trx.Blockid),
			BlockHeight: blockHeight,
			TxID:        hex.EncodeToString(trx.Txid),
			Decimal:     8,
			ConfirmTime: trx.Timestamp,
			Status:      openwallet.TxStatusSuccess,
			TxType:      txType,
			TxAction:    txAction,
		}
		wxID := openwallet.GenTransactionWxID(tx)
		tx.WxID = wxID
		extractData.Transaction = tx

		//bs.wm.Log.Debug("Transaction:", extractData.Transaction)
	}

	result.Success = success

}

//ExtractTxInput 提取交易单输入部分
func (bs *BlockScanner) extractTxInput(blockHeight uint64, trx *pb.Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanTargetFuncV2) ([]string, decimal.Decimal) {

	//vin := trx.Get("vin")

	var (
		from        = make([]string, 0)
		totalAmount = decimal.Zero
	)

	txType := uint64(0)

	createAt := time.Now().Unix()
	for i, output := range trx.TxInputs {

		//in := vin[i]

		txid := hex.EncodeToString(output.RefTxid)
		vout := output.RefOffset
		//
		//output, err := bs.wm.GetTxOut(txid, vout)
		//if err != nil {
		//	return err
		//}

		amount := common.BytesToDecimals(output.Amount, bs.wm.Decimal())
		addr := string(output.FromAddr)
		targetResult := scanAddressFunc(openwallet.ScanTargetParam{
			ScanTarget:     addr,
			Symbol:         bs.wm.Symbol(),
			ScanTargetType: openwallet.ScanTargetTypeAccountAddress})
		if targetResult.Exist {
			input := openwallet.TxInput{}
			input.SourceTxID = txid
			input.SourceIndex = uint64(vout)
			input.TxID = result.TxID
			input.Address = addr
			//transaction.AccountID = a.AccountID
			input.Amount = amount.String()
			input.Coin = openwallet.Coin{
				Symbol:     bs.wm.Symbol(),
				IsContract: false,
			}
			input.Index = uint64(i)
			input.Sid = openwallet.GenTxInputSID(txid, bs.wm.Symbol(), "", uint64(i))
			//input.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", result.TxID, i, addr))))
			input.CreateAt = createAt
			//在哪个区块高度时消费
			input.BlockHeight = blockHeight
			input.BlockHash = hex.EncodeToString(trx.Blockid)
			input.TxType = txType

			//transactions = append(transactions, &transaction)

			ed := result.extractData[targetResult.SourceKey]
			if ed == nil {
				ed = openwallet.NewBlockExtractData()
				result.extractData[targetResult.SourceKey] = ed
			}

			ed.TxInputs = append(ed.TxInputs, &input)

		}

		from = append(from, addr+":"+amount.String())
		totalAmount = totalAmount.Add(amount)

	}
	return from, totalAmount
}

//ExtractTxInput 提取交易单输入部分
func (bs *BlockScanner) extractTxOutput(blockHeight uint64, trx *pb.Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanTargetFuncV2) ([]string, decimal.Decimal) {

	var (
		to          = make([]string, 0)
		totalAmount = decimal.Zero
	)

	txType := uint64(0)

	vout := trx.TxOutputs
	txid := hex.EncodeToString(trx.Txid)
	//bs.wm.Log.Debug("vout:", vout.Array())
	createAt := time.Now().Unix()
	for i, output := range vout {

		amount := common.BytesToDecimals(output.Amount, bs.wm.Decimal())
		addr := string(output.ToAddr)
		targetResult := scanAddressFunc(openwallet.ScanTargetParam{
			ScanTarget:     addr,
			Symbol:         bs.wm.Symbol(),
			ScanTargetType: openwallet.ScanTargetTypeAccountAddress})
		if targetResult.Exist {

			//a := wallet.GetAddress(addr)
			//if a == nil {
			//	continue
			//}

			outPut := openwallet.TxOutPut{}
			outPut.TxID = txid
			outPut.Address = addr
			//transaction.AccountID = a.AccountID
			outPut.Amount = amount.String()
			outPut.Coin = openwallet.Coin{
				Symbol:     bs.wm.Symbol(),
				IsContract: false,
			}
			outPut.Index = uint64(i)
			outPut.Sid = openwallet.GenTxOutPutSID(txid, bs.wm.Symbol(), "", uint64(i))
			//outPut.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("output_%s_%d_%s", txid, n, addr))))

			outPut.CreateAt = createAt
			outPut.BlockHeight = blockHeight
			outPut.BlockHash = hex.EncodeToString(trx.Txid)
			outPut.TxType = txType

			//transactions = append(transactions, &transaction)

			ed := result.extractData[targetResult.SourceKey]
			if ed == nil {
				ed = openwallet.NewBlockExtractData()
				result.extractData[targetResult.SourceKey] = ed
			}

			ed.TxOutputs = append(ed.TxOutputs, &outPut)

		}

		to = append(to, addr+":"+amount.String())
		totalAmount = totalAmount.Add(amount)

	}

	return to, totalAmount
}


// extractSmartContractTransaction 提取智能合约交易单
func (bs *BlockScanner) extractSmartContractTransaction(blockHeight uint64, trx *pb.Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanTargetFuncV2) {

	for _, contractRequest := range trx.ContractRequests {

		moduleName := contractRequest.ModuleName
		contractName := contractRequest.ContractName
		contractAddress := moduleName + ":" + contractName

		//查找合约是否存在
		targetResult := scanAddressFunc(openwallet.ScanTargetParam{
			ScanTarget:     contractAddress,
			Symbol:         bs.wm.Symbol(),
			ScanTargetType: openwallet.ScanTargetTypeContractAddress})
		if !targetResult.Exist {
			continue
		}

		//查找合约对象信息
		contract, ok := targetResult.TargetInfo.(*openwallet.SmartContract)
		if !ok {
			bs.wm.Log.Errorf("tx to target result can not convert to openwallet.SmartContract")
			result.Success = false
			return
		}

		//没有纪录ABI，不处理提取
		if len(contract.GetABI()) == 0 {
			return
		}

		coin := openwallet.Coin{
			Symbol:     bs.wm.Symbol(),
			IsContract: true,
			ContractID: contract.ContractID,
			Contract:   *contract,
		}

		createAt := time.Now().Unix()

		//迭代每个日志，提取时间日志
		events := make([]*openwallet.SmartContractEvent, 0)
		for _, outPutExt := range trx.GetTxOutputsExt() {

			if string(outPutExt.Key) == EVENT_KEY {

				bucket := outPutExt.Bucket
				bucketContractAddress := moduleName + ":" + bucket

				logTargetResult := scanAddressFunc(openwallet.ScanTargetParam{
					ScanTarget:     bucketContractAddress,
					Symbol:         bs.wm.Symbol(),
					ScanTargetType: openwallet.ScanTargetTypeContractAddress})
				if !logTargetResult.Exist {
					continue
				}

				logContract, logOk := logTargetResult.TargetInfo.(*openwallet.SmartContract)
				if !logOk {
					bs.wm.Log.Errorf("log target result can not convert to openwallet.SmartContract")
					result.Success = false
					return
				}

				eventJSON := gjson.ParseBytes(outPutExt.Value)
				for _, e := range eventJSON.Array() {
					e := &openwallet.SmartContractEvent{
						Contract: logContract,
						Event:    e.Get("event").String(),
						Value:    e.Get("value").Raw,
					}

					events = append(events, e)
				}

			}
		}

		scReceipt := &openwallet.SmartContractReceipt{
			Coin:        coin,
			TxID:        hex.EncodeToString(trx.Txid),
			From:        trx.Initiator,
			To:          contractName,
			Fees:        "0",
			Value:       "0",
			RawReceipt:  "",
			Events:      events,
			BlockHash:   hex.EncodeToString(trx.Blockid),
			BlockHeight: blockHeight,
			ConfirmTime: createAt,
			Status:      "1",
			Reason:      "",
		}

		scReceipt.GenWxID()

		result.extractContractData[targetResult.SourceKey] = scReceipt

	}
}

//extractTokenTransfer 提取交易单中的代币交易
//func (bs *BlockScanner) extractTokenTransfer(blockHeight uint64, trx *pb.Transaction, result *ExtractResult, scanAddressFunc openwallet.BlockScanTargetFuncV2) {
//
//	var (
//		success = true
//	)
//
//	if trx == nil {
//		//记录哪个区块哪个交易单没有完成扫描
//		success = false
//	} else {
//
//		if trx.Isqrc20Transfer {
//			createAt := time.Now().Unix()
//			//QTUM交易单目前只允许包含一个代币交易
//			for _, tokenReceipt := range trx.TokenReceipts {
//
//				contractId := openwallet.GenContractID(bs.wm.Symbol(), tokenReceipt.ContractAddress)
//
//				coin := openwallet.Coin{
//					Symbol:     bs.wm.Symbol(),
//					IsContract: true,
//					ContractID: contractId,
//					Contract: openwallet.SmartContract{
//						ContractID: contractId,
//						Address:    tokenReceipt.ContractAddress,
//						Protocol:   "qrc20",
//						Symbol:     bs.wm.Symbol(),
//					},
//				}
//
//				targetResult := scanAddressFunc(openwallet.ScanTargetParam{
//					ScanTarget:     tokenReceipt.From,
//					Symbol:         bs.wm.Symbol(),
//					ScanTargetType: openwallet.ScanTargetTypeAccountAddress})
//				if targetResult.Exist {
//					input := openwallet.TxInput{}
//					input.TxID = trx.TxID
//					input.Address = tokenReceipt.From
//					//transaction.AccountID = a.AccountID
//					input.Amount = tokenReceipt.Amount
//					input.Coin = coin
//					input.Index = 0
//					input.Sid = openwallet.GenTxInputSID(tokenReceipt.TxHash, bs.wm.Symbol(), contractId, 0)
//					//input.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", result.TxID, i, addr))))
//					input.CreateAt = createAt
//					//在哪个区块高度时消费
//					input.BlockHeight = tokenReceipt.BlockHeight
//					input.BlockHash = tokenReceipt.BlockHash
//
//					//transactions = append(transactions, &transaction)
//
//					ed := result.extractContractData[targetResult.SourceKey]
//					if ed == nil {
//						ed = openwallet.NewBlockExtractData()
//						result.extractContractData[targetResult.SourceKey] = ed
//					}
//
//					ed.TxInputs = append(ed.TxInputs, &input)
//
//				}
//
//				targetResult2 := scanAddressFunc(openwallet.ScanTargetParam{
//					ScanTarget:     tokenReceipt.To,
//					Symbol:         bs.wm.Symbol(),
//					ScanTargetType: openwallet.ScanTargetTypeAccountAddress})
//				if targetResult2.Exist {
//					output := openwallet.TxOutPut{}
//					output.TxID = trx.TxID
//					output.Address = tokenReceipt.To
//					//transaction.AccountID = a.AccountID
//					output.Amount = tokenReceipt.Amount
//
//					output.Coin = coin
//					output.Index = 0
//					output.Sid = openwallet.GenTxOutPutSID(tokenReceipt.TxHash, bs.wm.Symbol(), contractId, 0)
//					//input.Sid = base64.StdEncoding.EncodeToString(crypto.SHA1([]byte(fmt.Sprintf("input_%s_%d_%s", result.TxID, i, addr))))
//					output.CreateAt = createAt
//					//在哪个区块高度时消费
//					output.BlockHeight = tokenReceipt.BlockHeight
//					output.BlockHash = tokenReceipt.BlockHash
//
//					//transactions = append(transactions, &transaction)
//
//					ed := result.extractContractData[targetResult2.SourceKey]
//					if ed == nil {
//						ed = openwallet.NewBlockExtractData()
//						result.extractContractData[targetResult2.SourceKey] = ed
//					}
//
//					ed.TxOutputs = append(ed.TxOutputs, &output)
//				}
//
//				blocktime := trx.Blocktime
//
//				for _, extractData := range result.extractContractData {
//					tx := &openwallet.Transaction{
//						From:        []string{tokenReceipt.From + ":" + tokenReceipt.Amount},
//						To:          []string{tokenReceipt.To + ":" + tokenReceipt.Amount},
//						Fees:        "0",
//						Coin:        coin,
//						BlockHash:   tokenReceipt.BlockHash,
//						BlockHeight: tokenReceipt.BlockHeight,
//						TxID:        tokenReceipt.TxHash,
//						Decimal:     0,
//						ConfirmTime: blocktime,
//						Status:      openwallet.TxStatusSuccess,
//						TxType:      0,
//					}
//					wxID := openwallet.GenTransactionWxID(tx)
//					tx.WxID = wxID
//					extractData.Transaction = tx
//
//					//bs.wm.Log.Debug("Transaction:", extractData.Transaction)
//				}
//			}
//		}
//
//		success = true
//
//	}
//
//	result.Success = success
//
//}

//newExtractDataNotify 发送通知
func (bs *BlockScanner) newExtractDataNotify(height uint64, extractData map[string]*openwallet.TxExtractData) error {

	for o, _ := range bs.Observers {
		for key, data := range extractData {
			err := o.BlockExtractDataNotify(key, data)
			if err != nil {
				bs.wm.Log.Error("BlockExtractDataNotify unexpected error:", err)
				//记录未扫区块
				unscanRecord := openwallet.NewUnscanRecord(height, "", "ExtractData Notify failed.", bs.wm.Symbol())
				err = bs.SaveUnscanRecord(unscanRecord)
				if err != nil {
					bs.wm.Log.Std.Error("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
				}

			}
		}
	}

	return nil
}

func (bs *BlockScanner) ExtractTransactionData(txid string, scanTargetFunc openwallet.BlockScanTargetFunc) (map[string][]*openwallet.TxExtractData, error) {

	scanTargetFuncV2 := func(target openwallet.ScanTargetParam) openwallet.ScanTargetResult {
		sourceKey, ok := scanTargetFunc(openwallet.ScanTarget{
			Address:          target.ScanTarget,
			Symbol:           bs.wm.Symbol(),
			BalanceModelType: bs.wm.BalanceModelType(),
		})
		return openwallet.ScanTargetResult{
			SourceKey: sourceKey,
			Exist:     ok,
		}
	}

	tx, err := bs.wm.RPC.QueryTx(txid)
	if err != nil {
		return nil, err
	}

	block, err := bs.wm.RPC.GetBlock(hex.EncodeToString(tx.Tx.Blockid))
	if err != nil {
		return nil, err
	}

	result := bs.ExtractTransaction(uint64(block.Height), hex.EncodeToString(tx.Tx.Blockid), tx.Tx, scanTargetFuncV2)
	if !result.Success {
		return nil, fmt.Errorf("extract transaction failed")
	}

	extData := make(map[string][]*openwallet.TxExtractData)
	for key, data := range result.extractData {
		txs := extData[key]
		if txs == nil {
			txs = make([]*openwallet.TxExtractData, 0)
		}
		txs = append(txs, data)
		extData[key] = txs
	}

	//for key, data := range result.extractContractData {
	//	txs := extData[key]
	//	if txs == nil {
	//		txs = make([]*openwallet.TxExtractData, 0)
	//	}
	//	txs = append(txs, data)
	//	extData[key] = txs
	//}

	return extData, nil
}


//ExtractTransactionAndReceiptData 提取交易单及交易回执数据
//@required
func (bs *BlockScanner) ExtractTransactionAndReceiptData(txid string, scanTargetFunc openwallet.BlockScanTargetFuncV2) (map[string][]*openwallet.TxExtractData, map[string]*openwallet.SmartContractReceipt, error) {

	tx, err := bs.wm.RPC.QueryTx(txid)
	if err != nil {
		bs.wm.Log.Errorf("get transaction by has failed, err: %v", err)
		return nil, nil, err
	}

	block, _ := bs.wm.RPC.GetBlock(hex.EncodeToString(tx.Tx.Blockid))
	if block == nil {
		return nil, nil, fmt.Errorf("get block failed")
	}
	result := bs.ExtractTransaction(uint64(block.Height), hex.EncodeToString(tx.Tx.Blockid), tx.Tx, scanTargetFunc)
	if !result.Success {
		return nil, nil, fmt.Errorf("extract transaction failed")
	}

	extData := make(map[string][]*openwallet.TxExtractData)
	for key, data := range result.extractData {
		txs := extData[key]
		if txs == nil {
			txs = make([]*openwallet.TxExtractData, 0)
		}
		txs = append(txs, data)
		extData[key] = txs
	}

	//for key, data := range result.extractContractData {
	//	txs := extData[key]
	//	if txs == nil {
	//		txs = make([]*openwallet.TxExtractData, 0)
	//	}
	//	txs = append(txs, data)
	//	extData[key] = txs
	//}

	return extData, result.extractContractData, nil
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

//GetScannedBlockHeader 获取当前已扫区块高度
func (bs *BlockScanner) GetScannedBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight      uint64 = 0
		blockChainStatus *pb.BCStatus
		hash             string
		err              error
	)

	blockHeight, hash, err = bs.GetLocalBlockHead()
	if err != nil {
		bs.wm.Log.Errorf("get local new block failed, err=%v", err)
		return nil, err
	}

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		blockChainStatus, err = bs.wm.RPC.GetBlockChainStatus()
		if err != nil {
			bs.wm.Log.Errorf("GetBlockChainStatus failed, err=%v", err)
			return nil, err
		}

		blockHeight = uint64(blockChainStatus.GetBlock().Height)

		//就上一个区块链为当前区块
		blockHeight = blockHeight - 1

		block, err := bs.wm.RPC.GetBlockByHeight(int64(blockHeight))
		if err != nil {
			bs.wm.Log.Errorf("get block spec by block number failed, err=%v", err)
			return nil, err
		}
		hash = hex.EncodeToString(block.GetBlockid())
	}

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}

//GetCurrentBlockHeader 获取当前区块高度
func (bs *BlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeight      uint64 = 0
		blockChainStatus *pb.BCStatus
		hash             string
		err              error
	)

	blockChainStatus, err = bs.wm.RPC.GetBlockChainStatus()
	if err != nil {
		bs.wm.Log.Errorf("GetBlockChainStatus failed, err=%v", err)
		return nil, err
	}

	block, err := bs.wm.RPC.GetBlockByHeight(blockChainStatus.GetBlock().Height)
	if err != nil {
		bs.wm.Log.Errorf("get block spec by block number failed, err=%v", err)
		return nil, err
	}
	hash = hex.EncodeToString(block.GetBlockid())

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}

func (bs *BlockScanner) GetBlockHeight() (uint64, error) {

	blockChainStatus, err := bs.wm.RPC.GetBlockChainStatus()
	if err != nil {
		bs.wm.Log.Errorf("GetBlockHeight failed, err: %v", err)
		return 0, err
	}
	return uint64(blockChainStatus.GetBlock().Height), nil
}

func (bs *BlockScanner) GetGlobalMaxBlockHeight() uint64 {
	height, _ := bs.GetBlockHeight()
	return height
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *BlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		return fmt.Errorf("block height to rescan must greater than 0 ")
	}

	block, err := bs.wm.RPC.GetBlockByHeight(int64(height))
	if err != nil {
		bs.wm.Log.Errorf("get block spec by block number[%v] failed, err=%v", height, err)
		return err
	}

	err = bs.SaveLocalBlockHead(height, hex.EncodeToString(block.Blockid))
	if err != nil {
		bs.wm.Log.Errorf("save local block scanned failed, err=%v", err)
		return err
	}

	return nil
}
