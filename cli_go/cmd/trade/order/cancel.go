// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package order

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"dex/internal/prompt"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

type cancelIn struct {
	contractAddress string
	walletAddress   string
	baseToken       common.Address
	quoteToken      common.Address
	orderID         *big.Int
}

type cancelOut struct {
	walletAddress string
	txHash        string
	minedBlock    uint64
}

func newCancelCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel an order you own",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCancel(cmd, cfg, ks)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().String("base", "", "Base token address or symbol")
	cmd.Flags().String("quote", "", "Quote token address or symbol")
	cmd.Flags().String("id", "", "Order id")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func runCancel(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	in, err := inputCancel(cmd, cfg, ks)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	out, err := processCancel(cmd, in, cfg, ks)
	if err != nil {
		return err
	}
	return outputCancel(cmd, out)
}

func inputCancel(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) (*cancelIn, error) {
	contractAddress, err := orderReadContractAddress(cmd, cfg.Get())
	if err != nil {
		return nil, err
	}
	walletAddress, _ := cmd.Flags().GetString("wallet")
	walletAddress, err = orderSelectWalletAddress(ks, strings.TrimSpace(walletAddress))
	if err != nil {
		return nil, err
	}
	baseToken, err := orderReadTokenAddressFlag(cmd, cfg, "base", "Base token (symbol or address)")
	if err != nil {
		return nil, err
	}
	quoteToken, err := orderReadTokenAddressFlag(cmd, cfg, "quote", "Quote token (symbol or address)")
	if err != nil {
		return nil, err
	}
	orderID, err := orderReadBigIntFlag(cmd, "id", "Order id", true)
	if err != nil {
		return nil, err
	}
	return &cancelIn{
		contractAddress: contractAddress,
		walletAddress:   walletAddress,
		baseToken:       baseToken,
		quoteToken:      quoteToken,
		orderID:         orderID,
	}, nil
}

func processCancel(cmd *cobra.Command, in *cancelIn, cfg *service.Service, ks *service.Keystore) (*cancelOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, orderRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()

	fmt.Fprintln(cmd.OutOrStdout(), "Review cancel")
	fmt.Fprintf(cmd.OutOrStdout(), "  Wallet:   %s\n", in.walletAddress)
	fmt.Fprintf(cmd.OutOrStdout(), "  Base:     %s\n", in.baseToken.Hex())
	fmt.Fprintf(cmd.OutOrStdout(), "  Quote:    %s\n", in.quoteToken.Hex())
	fmt.Fprintf(cmd.OutOrStdout(), "  Order ID: %s\n", in.orderID.String())
	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		ok, err := prompt.Confirm("Proceed and send cancel transaction")
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("aborted")
		}
	}

	walletService, err := service.NewWallet(ks, in.walletAddress)
	if err != nil {
		return nil, err
	}
	orderbookService, err := service.NewOrderBookService(rpcService, walletService, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	tx, err := orderbookService.CancelOrder(ctx, in.baseToken, in.quoteToken, in.orderID)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel tx submitted: %s\n", tx.Hash().Hex())
	fmt.Fprintln(cmd.OutOrStdout(), "Waiting for cancel transaction to be mined...")
	receipt, err := service.WaitForTxSuccess(ctx, rpcService, tx.Hash(), service.DefaultTxWaitTimeout)
	if err != nil {
		return nil, err
	}
	return &cancelOut{
		walletAddress: walletService.Address(),
		txHash:        tx.Hash().Hex(),
		minedBlock:    receipt.BlockNumber.Uint64(),
	}, nil
}

func outputCancel(cmd *cobra.Command, out *cancelOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", out.walletAddress)
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel tx: %s\n", out.txHash)
	fmt.Fprintf(cmd.OutOrStdout(), "Cancel mined in block: %d\n", out.minedBlock)
	return nil
}
