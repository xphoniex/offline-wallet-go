package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	hexMode = flag.String("raw", "", "hex of raw tx")
	verbose = flag.Bool("verbose", false, "print rlp parts while being decoded")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "-raw <data> [-verbose]")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Decodes a raw transaction and prints it in readable form.`)
	}
}

func main() {
	flag.Parse()

	var r io.Reader
	if *hexMode == "" {
		die("raw tx was not provided")
	}

	data, err := hex.DecodeString(strings.TrimPrefix(*hexMode, "0x"))
	if err != nil {
		die(err)
	}
	r = bytes.NewReader(data)

	out := os.Stdout
	err = rlpToText(r, out)
	if err != nil {
		die(err)
	}
}

func showTx(txType int, tx [][]byte) {
	fmt.Printf("tx: (type = %d)\n", txType)

	if txType == 0 {
		// rlp([nonce, gasPrice, gasLimit, to, value, data, v, r, s])
		fmt.Printf(" nonce \t\t\t\t= %d (0x%x)\n", bytesToBigInt(tx[0]), tx[0])
		fmt.Printf(" gas price\t\t\t= %9.9f GWei (0x%x)\n", weiTo(bytesToBigInt(tx[1]), params.GWei), tx[1])
		fmt.Printf(" gas limit\t\t\t= %d (0x%x)\n", bytesToBigInt(tx[2]), tx[2])
		fmt.Printf(" to\t\t\t\t= 0x%x\n", tx[3])
		fmt.Printf(" value\t\t\t\t= %18.18f ETH (0x%x)\n", weiTo(bytesToBigInt(tx[4]), params.Ether), tx[4])
		fmt.Printf(" data\t\t\t\t= 0x%x\n", tx[5])
	} else if txType == 2 {
		// https://eips.ethereum.org/EIPS/eip-1559
		// rlp([chain_id, nonce, max_priority_fee_per_gas, max_fee_per_gas, gas_limit, destination, amount, data, access_list, signature_y_parity, signature_r, signature_s])
		fmt.Printf(" chain id\t\t\t= %d (0x%x)\n", bytesToBigInt(tx[0]), tx[0])
		fmt.Printf(" nonce\t\t\t\t= %d (0x%x)\n", bytesToBigInt(tx[1]), tx[1])
		fmt.Printf(" max priority fee (per gas)\t= %9.9f GWei (0x%x)\n", weiTo(bytesToBigInt(tx[2]), params.GWei), tx[2])
		fmt.Printf(" max fee (per gas)\t\t= %9.9f GWei (0x%x)\n", weiTo(bytesToBigInt(tx[3]), params.GWei), tx[3])
		fmt.Printf(" gas limit\t\t\t= %d (0x%x)\n", bytesToBigInt(tx[4]), tx[4])
		fmt.Printf(" destination\t\t\t= 0x%x\n", tx[5])
		fmt.Printf(" amount\t\t\t\t= %18.18f ETH (0x%x)\n", weiTo(bytesToBigInt(tx[6]), params.Ether), tx[6])
		fmt.Printf(" data\t\t\t\t= 0x%x\n", tx[7])
	} else {
		die("tx type is invalid or not implemented:", txType)
	}
}

// https://github.com/ethereum/go-ethereum/issues/21221
func weiTo(wei *big.Int, target float64) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(target))
}

func bytesToBigInt(data []byte) (res *big.Int) {
	res = new(big.Int)
	res.SetBytes(data)
	return
}

func rlpToText(r io.Reader, out io.Writer) error {
	txType := 0

	s := rlp.NewStream(r, 0)
	for {
		res := make([][]byte, 0)
		err := dump(s, 0, out, &res)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		if *verbose {
			fmt.Fprintln(out)
		}
		if len(res) == 1 {
			txType = int(res[0][0])
		} else {
			showTx(txType, res)
		}
	}
	return nil
}

func dump(s *rlp.Stream, depth int, out io.Writer, res *[][]byte) error {
	kind, size, err := s.Kind()
	if err != nil {
		return err
	}
	switch kind {
	case rlp.Byte, rlp.String:
		str, err := s.Bytes()
		*res = append(*res, str)
		if err != nil {
			return err
		}
		if *verbose {
			if len(str) == 0 || isASCII(str) {
				fmt.Fprintf(out, "%s%q", ws(depth), str)
			} else {
				fmt.Fprintf(out, "%s%x", ws(depth), str)
			}
		}
		return nil
	case rlp.List:
		s.List()
		defer s.ListEnd()
		if size == 0 {
			if *verbose {
				fmt.Fprintf(out, ws(depth)+"[]")
			}
			*res = append(*res, []byte{})
			return nil
		} else {
			if *verbose {
				fmt.Fprintln(out, ws(depth)+"[")
			}
			for i := 0; ; i++ {
				if i > 0 && *verbose {
					fmt.Fprint(out, ",\n")
				}
				err := dump(s, depth+1, out, res)
				if err == rlp.EOL {
					break
				} else if err != nil {
					return err
				}
			}
			if *verbose {
				fmt.Fprint(out, ws(depth)+"]")
			}
		}
	}
	return nil
}

func isASCII(b []byte) bool {
	for _, c := range b {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}

func ws(n int) string {
	return strings.Repeat("  ", n)
}

func die(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
