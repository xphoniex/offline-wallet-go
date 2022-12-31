package main

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"reflect"
	"testing"
)

func TestPrivateWalletInitiation(t *testing.T) {
	cmd := "p/0x9d542624d9ef903daa81bfc3ba224ac15f3b55cd2bc5b09779b258d9fa753296"
	signer, err := initiateWalletFromCmd(&cmd)

	if err != nil {
		t.Error(err)
	}

	target := common.HexToAddress("0xfd94748666E47A0a1E6baE17eC72429390C66348")

	if *signer.address != target {
		t.Errorf("Signer address %s doesn't match target address %s", signer.address, target)
	}
}

func TestMnemonicWalletInitiation(t *testing.T) {
	cmd := "m/sound company scorpion ceiling museum edge keen diary bargain lake duty rapid"
	signer, err := initiateWalletFromCmd(&cmd)

	if err != nil {
		t.Error(err)
	}

	target := common.HexToAddress("0x0475F0d4a405A79b58f302BD22ECbdAF35B1759e")

	if *signer.address != target {
		t.Errorf("Signer address %s doesn't match target address %s", signer.address, target)
	}
}

func TestMnemonicWalletInitiationIndexed(t *testing.T) {
	cmd := "m/sound company scorpion ceiling museum edge keen diary bargain lake duty rapid/10"
	signer, err := initiateWalletFromCmd(&cmd)

	if err != nil {
		t.Error(err)
	}

	target := common.HexToAddress("0x81aD3882bF1DCBeFe6dEF162a51670a6993BF85f")

	if *signer.address != target {
		t.Errorf("Signer address %s doesn't match target address %s", signer.address, target)
	}
}

func TestEthTx(t *testing.T) {
	cmd := "p/0x9d542624d9ef903daa81bfc3ba224ac15f3b55cd2bc5b09779b258d9fa753296"
	signer, err := initiateWalletFromCmd(&cmd)

	if err != nil {
		t.Error(err)
	}

	cmd = "nonce=0 gas=21000 gasTipCap=1000000000 gasFeeCap=15000000000 chainID=1 to=0x0475F0d4a405A79b58f302BD22ECbdAF35B1759e amount=1000000000"
	target, _ := hex.DecodeString("02f86f0180843b9aca0085037e11d600825208940475f0d4a405a79b58f302bd22ecbdaf35b1759e843b9aca0080c001a0be8fc6f81a181ef6202f628e0380b044978411696bca5fb8b85c5279493a587ea029fbb50fd78aeccfd3192410760ac4359f7c022d0b80164e5f82ce3a786bc8f8")

	tx, err := cmdToTx(cmd, signer)
	if err != nil {
		t.Error(err)
	}

	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(raw, target) {
		t.Errorf("Received tx %s doesn't match target tx %s", hex.EncodeToString(raw), hex.EncodeToString(target))
	}
}

func TestERC20Tx(t *testing.T) {
	cmd := "p/0x9d542624d9ef903daa81bfc3ba224ac15f3b55cd2bc5b09779b258d9fa753296"
	signer, err := initiateWalletFromCmd(&cmd)

	if err != nil {
		t.Error(err)
	}

	cmd = "nonce=0 gas=78000 gasTipCap=1000000000 gasFeeCap=15000000000 chainID=1 to=0x0475F0d4a405A79b58f302BD22ECbdAF35B1759e amount=1000000000 token=usdc"
	target, _ := hex.DecodeString("02f8b10180843b9aca0085037e11d600830130b094a0b86991c6218b36c1d19d4a2e9eb0ce3606eb4880b844a9059cbb0000000000000000000000000475f0d4a405a79b58f302bd22ecbdaf35b1759e000000000000000000000000000000000000000000000000000000003b9aca00c080a04cf7c812a31b74273d51490821378c8aa753182df74dc46781a95b271f697598a071fa02edb760b383b99116d8b4c372a73121874564846eaf9f396ddba688570f")

	tx, err := cmdToTx(cmd, signer)
	if err != nil {
		t.Error(err)
	}

	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(raw, target) {
		t.Errorf("Received tx %s doesn't match target tx %s", hex.EncodeToString(raw), hex.EncodeToString(target))
	}
}

func TestGenericFnTx(t *testing.T) {
	cmd := "p/0x9d542624d9ef903daa81bfc3ba224ac15f3b55cd2bc5b09779b258d9fa753296"
	signer, err := initiateWalletFromCmd(&cmd)

	if err != nil {
		t.Error(err)
	}

	cmd = "nonce=0 gas=78000 gasTipCap=1000000000 gasFeeCap=15000000000 chainID=1 to=0x0475F0d4a405A79b58f302BD22ECbdAF35B1759e amount=1000000000 fn=hello()"
	target, _ := hex.DecodeString("02f8740180843b9aca0085037e11d600830130b0940475f0d4a405a79b58f302bd22ecbdaf35b1759e843b9aca008419ff1d21c080a01e01b4746b8038ecd4ad1321ee6b98a8d01cae95b8698634b9842b1f1cbe7c7ca069f35dc878b33114e019bad502032552709747131617304e6814f409face24bd")

	tx, err := cmdToTx(cmd, signer)
	if err != nil {
		t.Error(err)
	}

	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(raw, target) {
		t.Errorf("Received tx %s doesn't match target tx %s", hex.EncodeToString(raw), hex.EncodeToString(target))
	}

	cmd = "nonce=0 gas=78000 gasTipCap=1000000000 gasFeeCap=15000000000 chainID=1 to=0x0475F0d4a405A79b58f302BD22ECbdAF35B1759e amount=1000000000 data=19ff1d21"
	tx, err = cmdToTx(cmd, signer)
	if err != nil {
		t.Error(err)
	}

	raw, err = tx.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(raw, target) {
		t.Errorf("Received tx %s doesn't match target tx %s", hex.EncodeToString(raw), hex.EncodeToString(target))
	}

}
