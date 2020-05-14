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
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/shopspring/decimal"
	xupercom "github.com/xuperchain/xuper-sdk-go/common"
	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/utils"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
	"math/big"
	"strings"
	"time"
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return "", "", nil
}


//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return decoder.CreateSimpleRawTransaction(wrapper, rawTx, nil)
	}
	return nil
	//return decoder.CreateErc20TokenRawTransaction(wrapper, rawTx)
}

// CreateSummaryRawTransactionWithError 创建汇总交易，返回能原始交易单数组（包含带错误的原始交易单）
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {
	if !sumRawTx.Coin.IsContract {
		return decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	} else {
		return nil, nil
	}
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	var (
		rawTxWithErrArray []*openwallet.RawTransactionWithError
		rawTxArray        = make([]*openwallet.RawTransaction, 0)
		err               error
	)
	if !sumRawTx.Coin.IsContract {
		rawTxWithErrArray, err = decoder.CreateSimpleSummaryRawTransaction(wrapper, sumRawTx)
	} else {

	}
	if err != nil {
		return nil, err
	}
	for _, rawTxWithErr := range rawTxWithErrArray {
		if rawTxWithErr.Error != nil {
			continue
		}
		rawTxArray = append(rawTxArray, rawTxWithErr.RawTx)
	}
	return rawTxArray, nil
}

func (decoder *TransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction, tmpNonce *uint64) error {

	var (
		accountID    = rawTx.Account.AccountID
		usedUTXO     = make([]*pb.Utxo, 0)
		balance      = decimal.Zero
		totalSend    = decimal.Zero
		outputAddrs  = make(map[string]decimal.Decimal)
		destinations = make([]string, 0)
		limit        = 2000
		authAddrs    = make([]*openwallet.Address, 0)
	)

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, limit,
		"AccountID", accountID)
	if err != nil {
		return openwallet.NewError(openwallet.ErrAddressNotFound, err.Error())
	}

	if len(addresses) == 0 {
		return openwallet.Errorf(openwallet.ErrAccountNotAddress, "[%s] have not addresses", accountID)
	}

	//计算总发送金额
	for addr, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		totalSend = totalSend.Add(deamount)
		destinations = append(destinations, addr)
	}

	for _, addr := range addresses {

		unspents, err := decoder.wm.RPC.SelectUTXOBySize(addr.Address, true)
		if err != nil || unspents == nil {
			continue
		}

		//需要授权签名的地址列表
		authAddrs = append(authAddrs, addr)

		for _, u := range unspents {

			ua := common.BytesToDecimals(u.Amount, decoder.wm.Decimal())
			balance = balance.Add(ua)
			usedUTXO = append(usedUTXO, u)
			if balance.GreaterThanOrEqual(totalSend) {
				break
			}
		}
	}

	if balance.LessThan(totalSend) {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "The balance: %s is not enough! ", balance.String())
	}

	//取账户最后一个地址
	changeAddress := string(usedUTXO[0].ToAddr)
	changeAmount := balance.Sub(totalSend)

	decoder.wm.Log.Std.Notice("-----------------------------------------------")
	decoder.wm.Log.Std.Notice("From Account: %s", accountID)
	decoder.wm.Log.Std.Notice("To Address: %s", strings.Join(destinations, ", "))
	decoder.wm.Log.Std.Notice("Use: %v", balance.String())
	decoder.wm.Log.Std.Notice("Receive: %v", totalSend.String())
	decoder.wm.Log.Std.Notice("Change: %v", changeAmount.String())
	decoder.wm.Log.Std.Notice("Change Address: %v", changeAddress)
	decoder.wm.Log.Std.Notice("-----------------------------------------------")

	//装配输出
	for to, amount := range rawTx.To {
		decamount, _ := decimal.NewFromString(amount)
		outputAddrs = appendOutput(outputAddrs, to, decamount)
	}

	if changeAmount.GreaterThan(decimal.New(0, 0)) {
		outputAddrs = appendOutput(outputAddrs, changeAddress, changeAmount)
	}

	//最后创建交易单
	createTxErr := decoder.createRawTransaction(wrapper, rawTx, usedUTXO, authAddrs, outputAddrs)
	if createTxErr != nil {
		return createTxErr
	}

	return nil
}


//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//decoder.wm.Log.Std.Error("len of signatures error. ")
		return openwallet.Errorf(openwallet.ErrSignRawTransactionFailed, "transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		decoder.wm.Log.Error("get HDKey from wallet wrapper failed, err=%v", err)
		return err
	}

	for accountID, keySignatures := range rawTx.Signatures {

		decoder.wm.Log.Debug("accountID:", accountID)

		if keySignatures != nil {
			for _, keySignature := range keySignatures {

				msg, _ := hex.DecodeString(keySignature.Message)
				childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
				keyBytes, err := childKey.GetPrivateKeyBytes()
				if err != nil {
					return err
				}

				//签名交易
				signature, _, ret := owcrypt.Signature(keyBytes, nil, msg, decoder.wm.CurveType())
				if ret != owcrypt.SUCCESS {
					return fmt.Errorf("transaction hash signed failed")
				}

				keySignature.Signature = hex.EncodeToString(signature)
				//decoder.wm.Log.Debugf("signature: %s", hex.EncodeToString(signature))
			}
		}

		rawTx.Signatures[accountID] = keySignatures
	}

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//decoder.wm.Log.Std.Error("len of signatures error. ")
		return openwallet.Errorf(openwallet.ErrVerifyRawTransactionFailed, "transaction signature is empty")
	}

	if len(rawTx.RawHex) == 0 {
		return fmt.Errorf("transaction hex is empty")
	}

	rawHex := rawTx.RawHex
	var tx pb.Transaction
	err := json.Unmarshal([]byte(rawHex), &tx)
	if err != nil {
		return err
	}

	for accountID, keySignatures := range rawTx.Signatures {
		decoder.wm.Log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			compressPubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)
			msg, _ := hex.DecodeString(keySignature.Message)
			publickKey := owcrypt.PointDecompress(compressPubkey, decoder.wm.CurveType())

			if len(signature) != 64 {
				return fmt.Errorf("signature length is not equal to 32")
			}

			ret := owcrypt.Verify(publickKey[1:len(publickKey)], nil, msg, signature, decoder.wm.CurveType())
			if ret != owcrypt.SUCCESS {
				return fmt.Errorf("transaction verify signature failed: %s", keySignature.Signature)
			}

			pub := new(account.ECDSAPublicKey)
			pub.Curvname = elliptic.P256().Params().Name
			pub.X, pub.Y = elliptic.Unmarshal(elliptic.P256(), publickKey)
			pubJson, _ := json.Marshal(pub)

			r := new(big.Int)
			s := new(big.Int)
			derEncodeSig, err := utils.MarshalECDSASignature(r.SetBytes(signature[:32]), s.SetBytes(signature[32:]))
			if err != nil {
				return err
			}

			signInfo := &pb.SignatureInfo{
				PublicKey: string(pubJson),
				Sign:      derEncodeSig,
			}

			//填充交易发起者签名
			if keySignature.Address.Address == tx.Initiator {
				tx.InitiatorSigns = append(tx.InitiatorSigns, signInfo)
			}

			// 将签名填充进交易
			//tx.AuthRequireSigns = append(tx.AuthRequireSigns, signInfo)

			decoder.wm.Log.Debug("Signature:", keySignature.Signature)
			decoder.wm.Log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	txJSON, _ := json.Marshal(tx)
	rawTx.RawHex = string(txJSON)
	rawTx.IsCompleted = true

	return nil
}

// SubmitRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	rawHex := rawTx.RawHex
	var nTx pb.Transaction
	err := json.Unmarshal([]byte(rawHex), &nTx)
	if err != nil {
		return nil, err
	}

	// 最后一步，执行PostTx
	txid, err := decoder.wm.RPC.PostTx(&nTx)
	if err != nil {
		decoder.wm.Log.Errorf("raw: %s", rawHex)
		return nil, err
	}

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	decimals := int32(0)
	fees := "0"
	if rawTx.Coin.IsContract {
		decimals = int32(rawTx.Coin.Contract.Decimals)
		fees = "0"
	} else {
		decimals = int32(decoder.wm.Decimal())
		fees = rawTx.Fees
	}

	//记录一个交易单
	owtx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       fees,
		SubmitTime: time.Now().Unix(),
		TxType:     0,
	}

	owtx.WxID = openwallet.GenTransactionWxID(owtx)

	return owtx, nil
}

//CreateSimpleSummaryRawTransaction 创建汇总交易
func (decoder *TransactionDecoder) CreateSimpleSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		accountID      = sumRawTx.Account.AccountID
		minTransfer, _ = decimal.NewFromString(sumRawTx.MinTransfer)
		rawTxArray     = make([]*openwallet.RawTransactionWithError, 0)
		sumUnspents    = make([]*pb.Utxo, 0)
		outputAddrs    = make(map[string]decimal.Decimal, 0)
		sumAmount      = decimal.Zero
		authAddrs      = make([]*openwallet.Address, 0)
	)

	address, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit, "AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(address) == 0 {
		return nil, fmt.Errorf("[%s] have not addresses", accountID)
	}

	for i, addr := range address {

		addrBalance, err := decoder.wm.RPC.GetBalance(addr.Address)
		if err != nil {
			continue
		}

		//检查余额是否超过最低转账
		addrBalance_dec, _ := decimal.NewFromString(addrBalance.Balance)
		addrBalance_dec = addrBalance_dec.Shift(-decoder.wm.Decimal())
		if addrBalance_dec.LessThan(minTransfer) {
			continue
		}

		unspents, err := decoder.wm.RPC.SelectUTXOBySize(addr.Address, true)
		if err != nil {
			return nil, err
		}

		//尽可能筹够最大input数
		unspentLimit := decoder.wm.Config.MaxTxInputs - len(sumUnspents)
		if unspentLimit > 0 {
			if len(unspents) > unspentLimit {
				sumUnspents = append(sumUnspents, unspents[:unspentLimit]...)
			} else {
				sumUnspents = append(sumUnspents, unspents...)
			}
		}

		//需要授权签名的地址列表
		authAddrs = append(authAddrs, addr)

		//如果utxo已经超过最大输入，或遍历地址完结，就可以进行构建交易单
		if i == len(address)-1 || len(sumUnspents) >= decoder.wm.Config.MaxTxInputs {
			//执行构建交易单工作
			//计算这笔交易单的汇总数量
			for _, u := range sumUnspents {
				ua := common.BytesToDecimals(u.Amount, decoder.wm.Decimal())
				sumAmount = sumAmount.Add(ua)
			}

			/*
				汇总数量计算：

				1. 输入总数量 = 合计账户地址的所有utxo
				2. 账户地址输出总数量 = 账户地址保留余额 * 地址数
				3. 汇总数量 = 输入总数量 - 账户地址输出总数量 - 手续费
			*/

			decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

			if sumAmount.GreaterThan(decimal.Zero) {

				//最后填充汇总地址及汇总数量
				outputAddrs = appendOutput(outputAddrs, sumRawTx.SummaryAddress, sumAmount)
				//outputAddrs[sumRawTx.SummaryAddress] = sumAmount.StringFixed(decoder.wm.Decimal())

				raxTxTo := make(map[string]string, 0)
				for a, m := range outputAddrs {
					raxTxTo[a] = m.String()
				}

				//创建一笔交易单
				rawTx := &openwallet.RawTransaction{
					Coin:     sumRawTx.Coin,
					Account:  sumRawTx.Account,
					FeeRate:  sumRawTx.FeeRate,
					To:       raxTxTo,
					Fees:     "0",
					Required: 1,
				}

				createErr := decoder.createRawTransaction(wrapper, rawTx, sumUnspents, authAddrs, outputAddrs)
				rawTxWithErr := &openwallet.RawTransactionWithError{
					RawTx: rawTx,
					Error: openwallet.ConvertError(createErr),
				}

				//创建成功，添加到队列
				rawTxArray = append(rawTxArray, rawTxWithErr)

			}

			//清空临时变量
			sumUnspents = make([]*pb.Utxo, 0)
			authAddrs = make([]*openwallet.Address, 0)
			outputAddrs = make(map[string]decimal.Decimal, 0)
			sumAmount = decimal.Zero
		}
	}

	return rawTxArray, nil
}

//createRawTransaction 创建原始交易单
func (decoder *TransactionDecoder) createRawTransaction(
	wrapper openwallet.WalletDAI,
	rawTx *openwallet.RawTransaction,
	usedUTXO []*pb.Utxo,
	authAddrs []*openwallet.Address,
	to map[string]decimal.Decimal,
) error {

	var (
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		accountID        = rawTx.Account.AccountID
	)

	if len(usedUTXO) == 0 {
		return fmt.Errorf("utxo is empty")
	}

	if len(authAddrs) == 0 {
		return fmt.Errorf("authAddrs is empty")
	}

	if len(to) == 0 {
		return fmt.Errorf("Receiver addresses is empty! ")
	}

	//计算总发送金额
	for addr, amount := range to {
		//计算账户的实际转账amount
		addresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", accountID, "Address", addr)
		if findErr != nil || len(addresses) == 0 {
			accountTotalSent = accountTotalSent.Add(amount)
		}
	}

	// 声明一个交易，发起者为Alice地址，因为是转账，所以Desc字段什么都不填
	// 如果是提案等操作，将客户端的 --desc 参数写进去即可
	tx := &pb.Transaction{
		Version:   xupercom.TxVersion,
		Coinbase:  false,
		Desc:      []byte(""),
		Nonce:     global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: authAddrs[0].Address, //第一个地址作为发起者
	}
	// 填充交易的输入，即Select出来的Alice的utxo
	for _, utxo := range usedUTXO {
		txin := &pb.TxInput{
			RefTxid:   utxo.RefTxid,
			RefOffset: utxo.RefOffset,
			FromAddr:  utxo.ToAddr,
			Amount:    utxo.Amount,
		}
		tx.TxInputs = append(tx.TxInputs, txin)
		txFrom = append(txFrom, fmt.Sprintf("%s:%s", string(utxo.ToAddr), common.BytesToDecimals(utxo.Amount, decoder.wm.Decimal())))
	}

	for toAddr, toAmount := range to {
		// 填充交易的输出，即给Bob的utxo，注意Amount字段的类型
		amount := common.StringNumToBigIntWithExp(toAmount.String(), decoder.wm.Decimal())
		txout := &pb.TxOutput{
			ToAddr: []byte(toAddr),
			Amount: amount.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, txout)
		txTo = append(txTo, fmt.Sprintf("%s:%s", toAddr, toAmount.String()))
	}

	//for _, addr := range authAddrs {
	//	tx.AuthRequire = append(tx.AuthRequire, addr.Address)
	//}

	digestHash, dhErr := txhash.MakeTxDigestHash(tx)
	if dhErr != nil {
		return dhErr
	}

	//装配签名
	keySigs := make([]*openwallet.KeySignature, 0)

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	for _, addr := range authAddrs {

		signature := openwallet.KeySignature{
			EccType: decoder.wm.Config.CurveType,
			Address: addr,
			Message: hex.EncodeToString(digestHash),
		}

		keySigs = append(keySigs, &signature)

	}

	txJSON, _ := json.Marshal(tx)
	rawTx.RawHex = string(txJSON)

	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.Fees = "0"
	rawTx.Signatures[rawTx.Account.AccountID] = keySigs
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.String()
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil
}

func appendOutput(output map[string]decimal.Decimal, address string, amount decimal.Decimal) map[string]decimal.Decimal {
	if origin, ok := output[address]; ok {
		origin = origin.Add(amount)
		output[address] = origin
	} else {
		output[address] = amount
	}
	return output
}
