/*
 * Copyright 2019 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the confhope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */
package xuperchain

import (
	"github.com/blocktree/go-owcrypt"
)

const (
	Symbol    = "XUPER"
	CurveType = owcrypt.ECC_CURVE_NIST_P256
)

type ChainConfig struct {

	//币种
	Symbol string
	//钱包服务API
	ServerAPI string
	//曲线类型
	CurveType uint32
	//网络链名
	ChainName string
	//最大的输入数量
	MaxTxInputs int
}

func NewConfig(symbol string) *ChainConfig {
	c := ChainConfig{}
	c.Symbol = symbol
	c.CurveType = CurveType
	c.MaxTxInputs = 150
	return &c
}
