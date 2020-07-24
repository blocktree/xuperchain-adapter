package xuperchain

import (
	"crypto/elliptic"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/ethereum/go-ethereum/accounts/abi"
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

const (
	MODULE_XKERNEL = "xkernel"
	METHOD_DEPLOY  = "Deploy"
	EVENT_KEY = "com.github.blocktree.xcd.event"
)

type ContractDecoder struct {
	*openwallet.SmartContractDecoderBase
	wm *WalletManager
}

func (decoder *ContractDecoder) GetTokenBalanceByAddress(contract openwallet.SmartContract, address ...string) ([]*openwallet.TokenBalance, error) {

	tokenBalanceList := make([]*openwallet.TokenBalance, 0)

	return tokenBalanceList, nil
}

// PreInvokeContract 预执行合约
func (decoder *ContractDecoder) PreInvokeContract(wrapper openwallet.WalletDAI, rawTx *openwallet.SmartContractRawTransaction) (*pb.InvokeRPCRequest, *pb.PreExecWithSelectUTXOResponse, []*openwallet.Address, *openwallet.Error) {

	var (
		preSelUTXOReq *pb.PreExecWithSelectUTXORequest
		authAddrs     = make([]*openwallet.Address, 0)
	)
	if !rawTx.Coin.IsContract {
		return nil, nil, nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, "contract call msg invalid ")
	}

	abiJSON := rawTx.Coin.Contract.GetABI()
	if len(abiJSON) == 0 {
		return nil, nil, nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, "abi json is empty")
	}
	abiInstance, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, nil, nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, err.Error())
	}

	invokeRequest, encErr := decoder.wm.EncodeInvokeRequest(abiInstance, rawTx.Coin.Contract.Address, rawTx.ABIParam...)
	if encErr != nil {
		return nil, nil, nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, encErr.Error())
	}

	// generate preExe request
	var invokeRequests []*pb.InvokeRequest

	invokeRequests = append(invokeRequests, invokeRequest)

	invokeRPCReq := &pb.InvokeRPCRequest{
		Bcname:   decoder.wm.Config.ChainName,
		Requests: invokeRequests,
	}

	//系统内置的合约发布需要账户操作
	if invokeRequest.ModuleName == MODULE_XKERNEL && invokeRequest.MethodName == METHOD_DEPLOY {
		accountName := string(invokeRequest.Args["account_name"])
		acl, exist, findAccErr := decoder.wm.RPC.QueryACL(accountName)
		if findAccErr != nil {
			return nil, nil, nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, findAccErr.Error())
		}
		if !exist {
			return nil, nil, nil, openwallet.Errorf(openwallet.ErrAccountNotFound, "can not find account with name: %s", accountName)
		}

		//填充需要签名的地址
		for addr, _ := range acl.GetAcl().GetAksWeight() {
			invokeRPCReq.AuthRequire = append(invokeRPCReq.AuthRequire, accountName+"/"+addr)

			owAddress, findAddr := wrapper.GetAddress(addr)
			if findAddr != nil {
				return nil, nil, nil, openwallet.Errorf(openwallet.ErrAddressNotFound, "can not find address: %s", owAddress.Address)
			}

			authAddrs = append(authAddrs, owAddress)
		}
	}

	//账户的第一个地址为默认发起者
	defAddress, getErr := decoder.GetAssetsAccountDefAddress(wrapper, rawTx.Account.AccountID)
	if getErr != nil {
		return nil, nil, nil, getErr
	}

	invokeRPCReq.Initiator = defAddress.Address
	invokeRPCReq.AuthRequire = append(invokeRPCReq.AuthRequire, defAddress.Address)
	authAddrs = append(authAddrs, defAddress)

	//使用地址作为发起者的utxo
	preSelUTXOReq = &pb.PreExecWithSelectUTXORequest{
		Bcname:      decoder.wm.Config.ChainName,
		Address:     defAddress.Address,
		TotalAmount: 0,
		Request:     invokeRPCReq,
	}

	resp, preErr := decoder.wm.RPC.PreExecWithSelectUTXO(preSelUTXOReq)
	if preErr != nil {
		return nil, nil, nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, preErr.Error())
	}

	return invokeRPCReq, resp, authAddrs, nil
}

func (decoder *ContractDecoder) GetAssetsAccountDefAddress(wrapper openwallet.WalletDAI, accountID string) (*openwallet.Address, *openwallet.Error) {
	//获取wallet
	addresses, err := wrapper.GetAddressList(0, 1,
		"AccountID", accountID)
	if err != nil {
		return nil, openwallet.NewError(openwallet.ErrAddressNotFound, err.Error())
	}

	if len(addresses) == 0 {
		return nil, openwallet.Errorf(openwallet.ErrAccountNotAddress, "[%s] have not addresses", accountID)
	}
	return addresses[0], nil
}

//调用合约ABI方法
func (decoder *ContractDecoder) CallSmartContractABI(wrapper openwallet.WalletDAI, rawTx *openwallet.SmartContractRawTransaction) (*openwallet.SmartContractCallResult, *openwallet.Error) {

	if !rawTx.Coin.IsContract {
		return nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, "contract call msg invalid ")
	}

	abiJSON := rawTx.Coin.Contract.GetABI()
	if len(abiJSON) == 0 {
		return nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, "abi json is empty")
	}
	abiInstance, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, err.Error())
	}

	invokeRequest, encErr := decoder.wm.EncodeInvokeRequest(abiInstance, rawTx.Coin.Contract.Address, rawTx.ABIParam...)
	if encErr != nil {
		return nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, encErr.Error())
	}

	// generate preExe request
	var invokeRequests []*pb.InvokeRequest

	invokeRequests = append(invokeRequests, invokeRequest)

	invokeRPCReq := &pb.InvokeRPCRequest{
		Bcname:   decoder.wm.Config.ChainName,
		Requests: invokeRequests,
	}

	resp, preErr := decoder.wm.RPC.PreExec(invokeRPCReq)
	if preErr != nil {
		return nil, openwallet.Errorf(openwallet.ErrContractCallMsgInvalid, preErr.Error())
	}

	//gas := resp.GetResponse().GetGasUsed()
	//fmt.Printf("gas used: %v\n", gas)
	var rJson []byte
	for _, res := range resp.GetResponse().GetResponse() {
		//fmt.Printf("contract response: %s\n", string(res))
		rJson = res
	}

	callResult := &openwallet.SmartContractCallResult{
		Method: rawTx.ABIParam[0],
	}

	callResult.RawHex = hex.EncodeToString(rJson)
	callResult.Value = string(rJson)
	callResult.Status = openwallet.SmartContractCallResultStatusSuccess


	return callResult, nil
}

//创建原始交易单
func (decoder *ContractDecoder) CreateSmartContractRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.SmartContractRawTransaction) *openwallet.Error {

	invokeRPCReq, preResp, authAddrs, encErr := decoder.PreInvokeContract(wrapper, rawTx)
	if encErr != nil {
		return encErr
	}

	//手续费
	amount := big.NewInt(preResp.Response.GasUsed)

	// 构造一个发起的交易
	tx := &pb.Transaction{
		Version:   xupercom.TxVersion,
		Coinbase:  false,
		Desc:      []byte(""),
		Nonce:     global.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: invokeRPCReq.Initiator,
	}
	// 填充支付的手续费，手续费需要“转账”给地址“$”
	fee := &pb.TxOutput{
		ToAddr: []byte("$"),
		Amount: amount.Bytes(),
	}
	tx.TxOutputs = append(tx.TxOutputs, fee)
	// 填充select出来的utxo
	for _, utxo := range preResp.UtxoOutput.UtxoList {
		txin := &pb.TxInput{
			RefTxid:   utxo.RefTxid,
			RefOffset: utxo.RefOffset,
			FromAddr:  utxo.ToAddr,
			Amount:    utxo.Amount,
		}
		tx.TxInputs = append(tx.TxInputs, txin)
	}
	// 处理找零的逻辑
	total, _ := big.NewInt(0).SetString(preResp.UtxoOutput.TotalSelected, 10)
	if total.Cmp(amount) > 0 {
		delta := total.Sub(total, amount)
		charge := &pb.TxOutput{
			ToAddr: []byte(invokeRPCReq.Initiator),
			Amount: delta.Bytes(),
		}
		tx.TxOutputs = append(tx.TxOutputs, charge)
	}
	// 填充预执行的结果
	tx.ContractRequests = preResp.GetResponse().GetRequests()
	tx.TxInputsExt = preResp.GetResponse().GetInputs()
	tx.TxOutputsExt = preResp.GetResponse().GetOutputs()
	tx.AuthRequire = invokeRPCReq.AuthRequire

	digestHash, dhErr := txhash.MakeTxDigestHash(tx)
	if dhErr != nil {
		return openwallet.Errorf(openwallet.ErrCreateRawSmartContractTransactionFailed, dhErr.Error())
	}

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	for _, address := range authAddrs {

		signature := &openwallet.KeySignature{
			EccType: decoder.wm.Config.CurveType,
			Address: address,
			Message: hex.EncodeToString(digestHash),
		}

		keySigs := rawTx.Signatures[address.AccountID]
		if keySigs == nil {
			keySigs = make([]*openwallet.KeySignature, 0)
		}

		//装配签名
		keySigs = append(keySigs, signature)

		rawTx.Signatures[address.AccountID] = keySigs
	}

	txJSON, _ := json.Marshal(tx)
	rawTx.Raw = string(txJSON)
	rawTx.RawType = openwallet.TxRawTypeJSON
	rawTx.Fees = common.BigIntToDecimals(amount, decoder.wm.Decimal()).String()
	rawTx.IsBuilt = true
	rawTx.TxFrom = invokeRPCReq.Initiator
	rawTx.TxTo = rawTx.Coin.Contract.Address

	return nil
}

//SubmitRawTransaction 广播交易单
func (decoder *ContractDecoder) SubmitSmartContractRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.SmartContractRawTransaction) (*openwallet.SmartContractReceipt, *openwallet.Error) {

	nTx, err := decoder.VerifyRawTransaction(wrapper, rawTx)

	if !rawTx.IsCompleted {
		return nil, openwallet.Errorf(openwallet.ErrSubmitRawSmartContractTransactionFailed, "transaction is not completed validation")
	}

	// 最后一步，执行PostTx
	txid, err := decoder.wm.RPC.PostTx(nTx)
	if err != nil {
		decoder.wm.Log.Errorf("raw: %s", rawTx.Raw)
		return nil, openwallet.Errorf(openwallet.ErrSubmitRawSmartContractTransactionFailed, err.Error())
	}

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	owtx := &openwallet.SmartContractReceipt{
		Coin:  rawTx.Coin,
		TxID:  rawTx.TxID,
		From:  rawTx.TxFrom,
		To:    rawTx.TxTo,
		Value: rawTx.Value,
		Fees:  rawTx.Fees,
	}

	owtx.GenWxID()

	decoder.wm.Log.Infof("rawTx.AwaitResult = %v", rawTx.AwaitResult)
	//等待出块结果返回交易回执
	if rawTx.AwaitResult {
		bs := decoder.wm.GetBlockScanner()
		if bs == nil {
			decoder.wm.Log.Errorf("adapter blockscanner is nil")
			return owtx, nil
		}

		addrs := make(map[string]openwallet.ScanTargetResult)
		contract := &rawTx.Coin.Contract
		if contract == nil {
			decoder.wm.Log.Errorf("rawTx.Coin.Contract is nil")
			return owtx, nil
		}

		addrs[contract.Address] = openwallet.ScanTargetResult{SourceKey: contract.ContractID, Exist: true, TargetInfo: contract}

		scanTargetFunc := func(target openwallet.ScanTargetParam) openwallet.ScanTargetResult {
			result := addrs[target.ScanTarget]
			if result.Exist {
				return result
			}
			return openwallet.ScanTargetResult{SourceKey: "", Exist: false, TargetInfo: nil}
		}

		//默认超时90秒
		if rawTx.AwaitTimeout == 0 {
			rawTx.AwaitTimeout = 90
		}

		sleepSecond := 2 * time.Second

		//计算过期时间
		currentServerTime := time.Now()
		expiredTime := currentServerTime.Add(time.Duration(rawTx.AwaitTimeout) * time.Second)

		//等待交易单报块结果
		for {

			//当前重试时间
			currentReDoTime := time.Now()

			//decoder.wm.Log.Debugf("currentReDoTime = %s", currentReDoTime.String())
			//decoder.wm.Log.Debugf("expiredTime = %s", expiredTime.String())

			//超时终止
			if currentReDoTime.Unix() >= expiredTime.Unix() {
				break
			}

			_, contractResult, extractErr := bs.ExtractTransactionAndReceiptData(owtx.TxID, scanTargetFunc)
			if extractErr != nil {
				continue
				//decoder.wm.Log.Errorf("ExtractTransactionAndReceiptData failed, err: %v", extractErr)
				//return owtx, nil
			}

			//tx := txResult[contract.ContractID]
			receipt := contractResult[contract.ContractID]

			if receipt != nil {
				return receipt, nil
			}

			//等待sleepSecond秒重试
			time.Sleep(sleepSecond)
		}

	}

	return owtx, nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *ContractDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.SmartContractRawTransaction) (*pb.Transaction, error) {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//decoder.wm.Log.Std.Error("len of signatures error. ")
		return nil, openwallet.Errorf(openwallet.ErrVerifyRawTransactionFailed, "transaction signature is empty")
	}

	if len(rawTx.Raw) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	var (
		tx       pb.Transaction
		rawBytes []byte
	)

	//解析原始交易单
	switch rawTx.RawType {
	case openwallet.TxRawTypeHex:
		rawBytes, _ = hex.DecodeString(rawTx.Raw)
	case openwallet.TxRawTypeJSON:
		rawBytes = []byte(rawTx.Raw)
	case openwallet.TxRawTypeBase64:
		rawBytes, _ = base64.StdEncoding.DecodeString(rawTx.Raw)
	}

	err := json.Unmarshal(rawBytes, &tx)
	if err != nil {
		return nil, err
	}

	for accountID, keySignatures := range rawTx.Signatures {
		decoder.wm.Log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			compressPubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)
			msg, _ := hex.DecodeString(keySignature.Message)
			publickKey := owcrypt.PointDecompress(compressPubkey, decoder.wm.CurveType())

			if len(signature) != 64 {
				return nil, fmt.Errorf("signature length is not equal to 32")
			}

			ret := owcrypt.Verify(publickKey[1:len(publickKey)], nil, msg, signature, decoder.wm.CurveType())
			if ret != owcrypt.SUCCESS {
				return nil, fmt.Errorf("transaction verify signature failed: %s", keySignature.Signature)
			}

			pub := new(account.ECDSAPublicKey)
			pub.Curvname = elliptic.P256().Params().Name
			pub.X, pub.Y = elliptic.Unmarshal(elliptic.P256(), publickKey)
			pubJson, _ := json.Marshal(pub)

			r := new(big.Int)
			s := new(big.Int)
			derEncodeSig, err := utils.MarshalECDSASignature(r.SetBytes(signature[:32]), s.SetBytes(signature[32:]))
			if err != nil {
				return nil, err
			}

			signInfo := &pb.SignatureInfo{
				PublicKey: string(pubJson),
				Sign:      derEncodeSig,
			}

			if tx.Initiator == keySignature.Address.Address {
				tx.InitiatorSigns = append(tx.InitiatorSigns, signInfo)
			}

			// 将签名填充进交易
			tx.AuthRequireSigns = append(tx.AuthRequireSigns, signInfo)

			decoder.wm.Log.Debug("Signature:", keySignature.Signature)
			decoder.wm.Log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	txJSON, _ := json.Marshal(tx)
	rawTx.Raw = string(txJSON)
	rawTx.IsCompleted = true

	return &tx, nil
}
