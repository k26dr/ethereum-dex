// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package token

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"dex/internal/amount"
	"dex/internal/prompt"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func newSendCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send ERC20 tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			walletAddress, _ := cmd.Flags().GetString("wallet")
			walletAddress = strings.TrimSpace(walletAddress)
			cmd.SilenceUsage = true

			selectedWallet, err := tokenSelectWalletAddress(ks, walletAddress)
			if err != nil {
				return err
			}

			tokenAddress, err := tokenReadAddressFlag(cmd, "token", "Token contract address")
			if err != nil {
				return err
			}

			rpcService, err := service.NewRPC(cfg.Get().Network, tokenRPCConnectTimeout)
			if err != nil {
				return err
			}
			defer rpcService.Close()

			meta, err := cfg.ResolveTokenMetadata(context.Background(), rpcService, cfg.Get().Network.ChainID, tokenAddress)
			if err != nil {
				return err
			}

			walletAddressHex := common.HexToAddress(selectedWallet).Hex()
			readTokenService, err := service.NewTokenService(rpcService, nil, tokenAddress, cfg.Get().Network.ChainID)
			if err != nil {
				return err
			}

			balance, err := readTokenService.BalanceOf(context.Background(), common.HexToAddress(walletAddressHex))
			if err != nil {
				return err
			}
			if balance.Sign() <= 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", walletAddressHex)
				fmt.Fprintf(cmd.OutOrStdout(), "Token: %s\n", service.FormatTokenRef(meta.Symbol, tokenAddress.Hex()))
				fmt.Fprintf(cmd.OutOrStdout(), "Available: %s (raw: %s)\n", amount.FormatUnits(balance, meta.Decimals), balance.String())
				fmt.Fprintln(cmd.OutOrStdout(), "No tokens available in this wallet. Nothing to send.")
				return nil
			}

			sendAmount, err := tokenResolveSendAmount(cmd, meta.Decimals, balance)
			if err != nil {
				return err
			}

			toAddress, err := tokenReadAddressFlag(cmd, "to", "Recipient address")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Review send")
			fmt.Fprintf(cmd.OutOrStdout(), "  Wallet:    %s\n", walletAddressHex)
			fmt.Fprintf(cmd.OutOrStdout(), "  Token:     %s\n", service.FormatTokenRef(meta.Symbol, tokenAddress.Hex()))
			fmt.Fprintf(cmd.OutOrStdout(), "  Recipient: %s\n", toAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "  Available: %s (raw: %s)\n", amount.FormatUnits(balance, meta.Decimals), balance.String())
			fmt.Fprintf(cmd.OutOrStdout(), "  Amount:    %s (raw: %s)\n", amount.FormatUnits(sendAmount, meta.Decimals), sendAmount.String())
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				ok, err := prompt.Confirm("Proceed and send transfer transaction")
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("aborted")
				}
			}

			// Ask for password/decrypt only right before signing and sending.
			walletService, err := service.NewWallet(ks, walletAddressHex)
			if err != nil {
				return err
			}
			tokenService, err := service.NewTokenService(rpcService, walletService, tokenAddress, cfg.Get().Network.ChainID)
			if err != nil {
				return err
			}

			tx, err := tokenService.Send(context.Background(), toAddress, sendAmount)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Send tx submitted: %s\n", tx.Hash().Hex())
			fmt.Fprintln(cmd.OutOrStdout(), "Waiting for send transaction to be mined...")
			receipt, err := service.WaitForTxSuccess(context.Background(), rpcService, tx.Hash(), service.DefaultTxWaitTimeout)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", walletService.Address())
			fmt.Fprintf(cmd.OutOrStdout(), "Token: %s\n", service.FormatTokenRef(meta.Symbol, tokenAddress.Hex()))
			fmt.Fprintf(cmd.OutOrStdout(), "Recipient: %s\n", toAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Available: %s (raw: %s)\n", amount.FormatUnits(balance, meta.Decimals), balance.String())
			fmt.Fprintf(cmd.OutOrStdout(), "Amount: %s (raw: %s)\n", amount.FormatUnits(sendAmount, meta.Decimals), sendAmount.String())
			fmt.Fprintf(cmd.OutOrStdout(), "Send tx: %s\n", tx.Hash().Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Send mined in block: %d\n", receipt.BlockNumber.Uint64())
			return nil
		},
	}

	cmd.Flags().String("token", "", "Token contract address")
	cmd.Flags().String("to", "", "Recipient address")
	cmd.Flags().String("amount", "", "Token amount (decimal token units, e.g. 1.5)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func tokenResolveSendAmount(cmd *cobra.Command, decimals uint8, available *big.Int) (*big.Int, error) {
	amountRaw, _ := cmd.Flags().GetString("amount")
	amountRaw = strings.TrimSpace(amountRaw)

	if amountRaw == "" {
		promptLabel := fmt.Sprintf(
			"Send amount (available: %s, type 'all' to send full balance)",
			amount.FormatUnits(available, decimals),
		)
		value, err := tokenReadAmountFlag(cmd, "amount", promptLabel)
		if err != nil {
			return nil, err
		}
		amountRaw = strings.TrimSpace(value)
	}

	if strings.EqualFold(amountRaw, "all") {
		if available == nil || available.Sign() <= 0 {
			return nil, fmt.Errorf("no available balance to send")
		}
		return new(big.Int).Set(available), nil
	}

	sendAmount, err := amount.ParseUnits(amountRaw, decimals)
	if err != nil {
		return nil, fmt.Errorf("invalid send amount %q: %w", amountRaw, err)
	}
	if sendAmount.Sign() <= 0 {
		return nil, fmt.Errorf("send amount must be greater than zero")
	}
	if available != nil && sendAmount.Cmp(available) > 0 {
		return nil, fmt.Errorf(
			"send amount exceeds available balance (%s)",
			amount.FormatUnits(available, decimals),
		)
	}
	return sendAmount, nil
}
