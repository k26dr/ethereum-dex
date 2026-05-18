// Copyright © 2026 0xTrooper, Kedar Iyer (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package test

import (
	"dex/service"
	"github.com/spf13/cobra"
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func NewCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests",
		Long: "Run tests against a simulated testnet",
		Run: func(cmd *cobra.Command, args []string) {
			//StartBlockchain()
		},
	}

	return cmd
}


func StartBlockchain () {
    privateKey, err := crypto.GenerateKey()
    if err != nil {
        log.Fatal(err)
    }

    auth := bind.NewKeyedTransactor(privateKey)

    balance := new(big.Int)
    balance.SetString("10000000000000000000", 10) // 10 eth in wei

    address := auth.From
    genesisAlloc := map[common.Address]core.GenesisAccount{
        address: {
            Balance: balance,
        },
    }

    blockGasLimit := uint64(4712388)
    client := backends.NewSimulatedBackend(genesisAlloc, blockGasLimit)

    fromAddress := auth.From
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        log.Fatal(err)
    }

    value := big.NewInt(1000000000000000000) // in wei (1 eth)
    gasLimit := uint64(21000)                // in units
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
    var data []byte
    tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
    signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
    if err != nil {
        log.Fatal(err)
    }

    err = client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex()) // tx sent: 0xec3ceb05642c61d33fa6c951b54080d1953ac8227be81e7b5e4e2cfed69eeb51

    client.Commit()

    receipt, err := client.TransactionReceipt(context.Background(), signedTx.Hash())
    if err != nil {
        log.Fatal(err)
    }
    if receipt == nil {
        log.Fatal("receipt is nil. Forgot to commit?")
    }

    fmt.Printf("status: %v\n", receipt.Status) // status: 1
}
