package main

import (
	"github.com/davecgh/go-spew/spew"
)

func deraw() {

	// eip-1559 tx: 0xc2163f50770bd4bfd3c13848b405a56a451ae2a39cfa5a236ea2738ce44aa9df
	rawTx := "02f8740181bf8459682f00851191460ee38252089497e542ec6b81dea28f212775ce8ac436ab77a7df880de0b6b3a764000080c080a02bc11202cee115fe22558ce2edb25c621266ce75f75e9b10da9a2ae72460ad4ea07d573eef31fdebf0f5f93eb7721924a082907419eb97a8dda0dd20a4a5b954a1"

	// legacy tx
	//rawTx := "f86d8202b28477359400825208944592d8f8d7b001e72cb26a73e4fa1806a51ac79d880de0b6b3a7640000802ca05924bde7ef10aa88db9c66dd4f5fb16b46dff2319b9968be983118b57bb50562a001b24b31010004f13d9a26b320845257a6cfc2bf819a3d55e3fc86263c5f0772"

	tx := &types.Transaction{}
	rawTxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		fmt.Println("err:", err)
		return 1
	}

	err = tx.UnmarshalBinary(rawTxBytes)
	if err != nil {
		fmt.Println("err:", err)
		return 1
	}

	spew.Dump(tx)
}
