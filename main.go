package main

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

func main() {
	var cmd string
	scanner := bufio.NewScanner(os.Stdin)

START:
	fmt.Printf("%% Create your wallet:\n> ")
	for scanner.Scan() {
		cmd = scanner.Text()
		break
	}

	if cmd == "restart" {
		goto START
	}

	signer, err := initiateWalletFromCmd(&cmd)
	if err != nil {
		fmt.Println("err:", err)
		goto START
	}
	fmt.Println("Opened Wallet:", signer.address)

OPERATION:
	fmt.Printf("%% Operation:\n(nonce, gas, gasTipCap, gasFeeCap, chainID, to, amount, token)\n> ")
POST_OPERATION_PROMPT:
	for scanner.Scan() {
		cmd = scanner.Text()

		if cmd == "restart" {
			goto START
		}

		if cmd == "" {
			fmt.Printf("> ")
			goto POST_OPERATION_PROMPT
		}

		tx, err := cmdToTx(cmd, signer)
		if err != nil {
			fmt.Println("err:", err)
			goto OPERATION
		}

		raw, err := tx.MarshalBinary()
		if err != nil {
			fmt.Println("err:", err)
			goto OPERATION
		}

		fmt.Printf("\n0x%x\n\n", raw)
		goto OPERATION
	}
}

func initiateWalletFromCmd(cmd *string) (*signer, error) {
	var privateKey string

	if strings.HasPrefix(*cmd, "p/") {
		privateKey = strings.TrimPrefix(*cmd, "p/")
		privateKey = strings.TrimPrefix(privateKey, "0x")
	} else if strings.HasPrefix(*cmd, "f/") {
		privateKeyFile := strings.TrimPrefix(*cmd, "f/")
		privateKey_, err := ioutil.ReadFile(privateKeyFile)
		if err != nil {
			return nil, err
		}
		privateKey = strings.TrimSuffix(string(privateKey_), "\n")
	} else if strings.HasPrefix(*cmd, "m/") {
		mnemonic := strings.TrimPrefix(*cmd, "m/")
		index := 0
		pieces := strings.Split(mnemonic, "/")
		if len(pieces) == 2 {
			idx, err := strconv.Atoi(pieces[1])
			if err != nil {
				return nil, err
			}

			index = idx
			mnemonic = pieces[0]
		}

		wallet, err := hdwallet.NewFromMnemonic(mnemonic)
		if err != nil {
			return nil, err
		}

		path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", index))
		account, err := wallet.Derive(path, true)
		if err != nil {
			return nil, err
		}

		privateKey, err = wallet.PrivateKeyHex(account)
		if err != nil {
			return nil, err
		}

		if index != 0 {
			fmt.Println("Index:", index, " Private key:", privateKey)
		}
	}

	signer, err := signerFromKey(privateKey)
	return signer, err
}

func cmdToTx(cmd string, signer *signer) (*types.Transaction, error) {
	var nonce uint64
	var gas uint64
	var to common.Address
	var data []byte

	gasTipCap := new(big.Int)
	gasFeeCap := new(big.Int)
	chainID := new(big.Int)
	amount := new(big.Int)

	var token string
	var tokenAddress common.Address

	for _, part := range strings.Split(cmd, " ") {
		var err error
		var ok bool

		keyValue := strings.Split(part, "=")
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("use format key=value")
		}
		key := keyValue[0]
		value := keyValue[1]

		switch key {
		case "nonce":
			nonce, err = strconv.ParseUint(value, 10, 64)
		case "gas":
			gas, err = strconv.ParseUint(value, 10, 64)
		case "gasTipCap":
			gasTipCap, ok = gasTipCap.SetString(value, 10)
			if !ok {
				err = fmt.Errorf("invalid value for gasTipCap")
			}
		case "gasFeeCap":
			gasFeeCap, ok = gasFeeCap.SetString(value, 10)
			if !ok {
				err = fmt.Errorf("invalid value for gasFeeCap")
			}
		case "chainID":
			chainID, ok = chainID.SetString(value, 10)
			if !ok {
				err = fmt.Errorf("invalid value for chainID")
			}
		case "to":
			to = common.HexToAddress(value)
		case "amount":
			amount, ok = amount.SetString(value, 10)
			if !ok {
				err = fmt.Errorf("invalid value for amount")
			}
		case "data":
			if value != "" && value != "0x" {
				data, err = hex.DecodeString(value)
			}
		case "token":
			token = value
		}

		if err != nil {
			return nil, err
		}
	}

	if token != "" {
		if token == "dai" {
			tokenAddress = common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
		} else if token == "rdai" {
			tokenAddress = common.HexToAddress("0xad6d458402f60fd3bd25163575031acdce07538d")
		} else if token == "usdc" {
			tokenAddress = common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
		} else if token == "usdt" {
			tokenAddress = common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7")
		} else {
			tokenAddress = common.HexToAddress(token)
		}
	}

	hasEthTxFields := gas != 0 && gasTipCap != nil && gasFeeCap != nil && chainID != nil && len(to.Bytes()) != 0 && amount != nil
	if !hasEthTxFields {
		return nil, fmt.Errorf("need all fields: nonce, gas, gasTipCap, gasFeeCap, chainID, to, amount")
	}

	var tx *types.Transaction
	var err error

	if token != "" {
		data = transferERC20(tokenAddress, to, amount)
		amount = big.NewInt(0)
		tx, err = signer.createDynamicFeeTx(nonce, gas, gasTipCap, gasFeeCap, chainID, amount, data, &tokenAddress)
	} else {
		tx, err = signer.createDynamicFeeTx(nonce, gas, gasTipCap, gasFeeCap, chainID, amount, data, &to)
	}

	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(tx.ChainId()), signer.privateKey)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func transferERC20(tokenAddress, toAddress common.Address, amount *big.Int) (data []byte) {
	fnSig := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(fnSig)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	return
}

type signer struct {
	privateKey *ecdsa.PrivateKey
	address    *common.Address
}

func signerFromKey(privateHex string) (*signer, error) {
	var err error
	var privateKey *ecdsa.PrivateKey

	if privateKey, err = crypto.HexToECDSA(privateHex); err != nil {
		return nil, err
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return &signer{
		privateKey,
		&address,
	}, nil
}

func (s *signer) createDynamicFeeTx(nonce, gas uint64, gasTipCap, gasFeeCap, chainID, amount *big.Int, data []byte, toAddress *common.Address) (*types.Transaction, error) {
	txdata := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        toAddress,
		Value:     amount,
		Gas:       gas,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Data:      data,
	}

	return types.NewTx(txdata), nil
}
