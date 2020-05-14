package xuperchain_addrdec

import (
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/blocktree/go-owcrypt"
	"github.com/blocktree/openwallet/v2/openwallet"
)

var (
	alphabet = addressEncoder.BTCAlphabet
)

var (
	Nist   = addressEncoder.AddressType{"base58", alphabet, "doubleSHA256", "h160", 20, []byte{0x01}, nil}
	Gm     = addressEncoder.AddressType{"base58", alphabet, "doubleSHA256", "h160", 20, []byte{0x02}, nil}
	NistSN = addressEncoder.AddressType{"base58", alphabet, "doubleSHA256", "h160", 20, []byte{0x03}, nil}
)

//AddressDecoderV2
type AddressDecoderV2 struct {
	*openwallet.AddressDecoderV2Base
	eccType uint32
}

//NewAddressDecoder 地址解析器
func NewAddressDecoder(eccType uint32) *AddressDecoderV2 {
	decoder := AddressDecoderV2{eccType: eccType}
	return &decoder
}

//AddressDecode 地址解析
func (dec *AddressDecoderV2) AddressDecode(addr string, opts ...interface{}) ([]byte, error) {

	cfg := Nist

	if len(opts) > 0 {
		for _, opt := range opts {
			if at, ok := opt.(addressEncoder.AddressType); ok {
				cfg = at
			}
		}
	}

	return addressEncoder.AddressDecode(addr, cfg)
}

//AddressEncode 地址编码
func (dec *AddressDecoderV2) AddressEncode(hash []byte, opts ...interface{}) (string, error) {

	cfg := Nist

	if len(opts) > 0 {
		for _, opt := range opts {
			if at, ok := opt.(addressEncoder.AddressType); ok {
				cfg = at
			}
		}
	}

	//如果是压缩公钥，进行解压
	if len(hash) == 33 {
		hash = owcrypt.PointDecompress(hash, dec.eccType)
	}

	address := addressEncoder.AddressEncode(hash, cfg)

	return address, nil
}

// AddressVerify 地址校验
func (dec *AddressDecoderV2) AddressVerify(address string, opts ...interface{}) bool {

	_, err := dec.AddressDecode(address, opts)
	if err != nil {
		return false
	}

	return true
}
