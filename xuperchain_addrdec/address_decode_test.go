package xuperchain_addrdec

import (
	"encoding/hex"
	"github.com/blocktree/go-owcrypt"
	"testing"
)

func TestAddressDecoder_AddressEncode(t *testing.T) {

	expect := "nofJPPzVCpDnXixVhLWfEeyzgDDAu9rSo"

	addrdec := NewAddressDecoder(owcrypt.ECC_CURVE_NIST_P256)
	hash, _ := hex.DecodeString("04ed47738bb56aabfc01e0d962b8250b6b0a4f484438daa936a03787f4846c42ad0c426c0beac8bd3cceca8163563b90b35a08a82142ee02e698f5a3b4131b5f9b")
	addr, _ := addrdec.AddressEncode(hash)

	if expect != addr {
		t.Errorf("address: %s is not equal to expect: %s", addr, expect)
	}

}

func TestAddressDecoder_AddressDecode(t *testing.T) {

	expect := "f670c3fa9d6ba96157e1ada7413f3f83a1d2e2a9"

	addrdec := NewAddressDecoder(owcrypt.ECC_CURVE_NIST_P256)
	addr := "nofJPPzVCpDnXixVhLWfEeyzgDDAu9rSo"
	hash, _ := addrdec.AddressDecode(addr)
	t.Logf("hash: %s", hex.EncodeToString(hash))

	hashHex := hex.EncodeToString(hash)
	if expect != hashHex {
		t.Errorf("address: %s is not equal to expect: %s", hashHex, expect)
	}

}

func TestAddressDecoderV2_AddressVerify(t *testing.T) {
	addrdec := NewAddressDecoder(owcrypt.ECC_CURVE_NIST_P256)
	expect := true
	addr := "nofJPPzVCpDnXixVhLWfEeyzgDDAu9rSo"


	valid := addrdec.AddressVerify(addr)

	if valid != expect {
		t.Errorf("Failed to verify %s valid address", addr)
	}

}