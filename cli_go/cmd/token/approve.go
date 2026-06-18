// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package token

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"dex/internal/amount"
	"dex/internal/prompt"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const tokenRPCConnectTimeout = 15 * time.Second

func newApproveCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve token allowance for a spender",
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

			spenderAddress, usedDefaultSpender, err := tokenReadSpenderAddress(cmd, cfg.Get())
			if err != nil {
				return err
			}
			if usedDefaultSpender {
				fmt.Fprintf(cmd.OutOrStdout(), "No spender was provided. Using default exchange contract from config: %s\n", spenderAddress.Hex())
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

			readTokenService, err := service.NewTokenService(rpcService, nil, tokenAddress, cfg.Get().Network.ChainID)
			if err != nil {
				return err
			}

			approveAmount, err := tokenResolveApproveAmount(cmd, readTokenService, selectedWallet, meta.Decimals)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Review approve")
			fmt.Fprintf(cmd.OutOrStdout(), "  Wallet:  %s\n", selectedWallet)
			fmt.Fprintf(cmd.OutOrStdout(), "  Token:   %s\n", service.FormatTokenRef(meta.Symbol, tokenAddress.Hex()))
			fmt.Fprintf(cmd.OutOrStdout(), "  Spender: %s\n", spenderAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "  Amount:  %s (raw: %s)\n", amount.FormatUnits(approveAmount, meta.Decimals), approveAmount.String())
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				ok, err := prompt.Confirm("Proceed and send approve transaction")
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("aborted")
				}
			}

			walletService, err := service.NewWallet(ks, selectedWallet)
			if err != nil {
				return err
			}

			tokenService, err := service.NewTokenService(rpcService, walletService, tokenAddress, cfg.Get().Network.ChainID)
			if err != nil {
				return err
			}

			tx, err := tokenService.Approve(context.Background(), spenderAddress, approveAmount)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Approve tx submitted: %s\n", tx.Hash().Hex())
			fmt.Fprintln(cmd.OutOrStdout(), "Waiting for approve transaction to be mined...")
			receipt, err := service.WaitForTxSuccess(context.Background(), rpcService, tx.Hash(), service.DefaultTxWaitTimeout)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Signer wallet: %s\n", walletService.Address())
			fmt.Fprintf(cmd.OutOrStdout(), "Token: %s\n", service.FormatTokenRef(meta.Symbol, tokenAddress.Hex()))
			fmt.Fprintf(cmd.OutOrStdout(), "Spender: %s\n", spenderAddress.Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Amount: %s (raw: %s)\n", amount.FormatUnits(approveAmount, meta.Decimals), approveAmount.String())
			fmt.Fprintf(cmd.OutOrStdout(), "Approve tx: %s\n", tx.Hash().Hex())
			fmt.Fprintf(cmd.OutOrStdout(), "Approve mined in block: %d\n", receipt.BlockNumber.Uint64())
			return nil
		},
	}

	cmd.Flags().String("token", "", "Token contract address")
	cmd.Flags().String("spender", "", "Spender address (defaults to config.contract.address)")
	cmd.Flags().String("amount", "", "Allowance amount (decimal token units, e.g. 1.5)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func tokenReadSpenderAddress(cmd *cobra.Command, cfg *service.Config) (common.Address, bool, error) {
	spenderValue, _ := cmd.Flags().GetString("spender")
	spenderValue = strings.TrimSpace(spenderValue)
	if spenderValue != "" {
		if !common.IsHexAddress(spenderValue) {
			return common.Address{}, false, fmt.Errorf("invalid spender address %q", spenderValue)
		}
		return common.HexToAddress(spenderValue), false, nil
	}

	if cfg != nil {
		configAddress := strings.TrimSpace(cfg.Contract.Address)
		if common.IsHexAddress(configAddress) {
			spender := common.HexToAddress(configAddress)
			if spender != (common.Address{}) {
				return spender, true, nil
			}
		}
	}

	spender, err := tokenReadAddressPrompt("Spender address")
	if err != nil {
		return common.Address{}, false, err
	}
	return spender, false, nil
}

func tokenReadAddressFlag(cmd *cobra.Command, flagName string, promptLabel string) (common.Address, error) {
	value, _ := cmd.Flags().GetString(flagName)
	value = strings.TrimSpace(value)
	if value == "" {
		return tokenReadAddressPrompt(promptLabel)
	}
	if !common.IsHexAddress(value) {
		return common.Address{}, fmt.Errorf("invalid %s %q", flagName, value)
	}
	return common.HexToAddress(value), nil
}

func tokenReadAddressPrompt(label string) (common.Address, error) {
	for {
		value, err := tokenAsk(label)
		if err != nil {
			return common.Address{}, err
		}
		value = strings.TrimSpace(value)
		if !common.IsHexAddress(value) {
			fmt.Fprintf(os.Stderr, "Invalid address %q.\n", value)
			continue
		}
		return common.HexToAddress(value), nil
	}
}

func tokenReadAmountFlag(cmd *cobra.Command, flagName string, promptLabel string) (string, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			next, err := tokenAsk(promptLabel)
			if err != nil {
				return "", err
			}
			value = strings.TrimSpace(next)
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return "", err
			}
		}
		return value, nil
	}
}

func tokenAsk(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("aborted")
	}
	return scanner.Text(), nil
}

func tokenSelectWalletAddress(ks *service.Keystore, requestedAddress string) (string, error) {
	if ks == nil {
		return "", fmt.Errorf("keystore service is not initialized")
	}
	wallets := ks.List()
	if len(wallets) == 0 {
		return "", fmt.Errorf("no wallets in keystore")
	}

	if strings.TrimSpace(requestedAddress) != "" {
		if !common.IsHexAddress(requestedAddress) {
			return "", fmt.Errorf("invalid wallet address %q", requestedAddress)
		}
		wanted := common.HexToAddress(requestedAddress).Hex()
		for _, wallet := range wallets {
			if strings.EqualFold(wallet, wanted) {
				return wanted, nil
			}
		}
		return "", fmt.Errorf("wallet %s not found in keystore", wanted)
	}

	if len(wallets) == 1 {
		return wallets[0], nil
	}

	for i, wallet := range wallets {
		fmt.Fprintf(os.Stderr, "[%d] %s\n", i+1, wallet)
	}
	fmt.Fprintln(os.Stderr)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "Select wallet [1-%d]: ", len(wallets))
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("aborted")
		}
		n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || n < 1 || n > len(wallets) {
			fmt.Fprintf(os.Stderr, "Please enter a number between 1 and %d.\n", len(wallets))
			continue
		}
		return wallets[n-1], nil
	}
}

func tokenResolveApproveAmount(cmd *cobra.Command, tokenService *service.TokenService, walletAddress string, decimals uint8) (*big.Int, error) {
	amountRaw, _ := cmd.Flags().GetString("amount")
	if strings.TrimSpace(amountRaw) != "" {
		value, err := tokenReadAmountFlag(cmd, "amount", "Approve amount")
		if err != nil {
			return nil, err
		}
		parsed, err := amount.ParseUnits(value, decimals)
		if err != nil {
			return nil, fmt.Errorf("invalid approve amount %q: %w", value, err)
		}
		if parsed.Sign() <= 0 {
			return nil, fmt.Errorf("approve amount must be greater than zero")
		}
		return parsed, nil
	}

	balance, err := tokenService.BalanceOf(context.Background(), common.HexToAddress(walletAddress))
	if err != nil {
		return nil, err
	}

	return prompt.SelectAllowanceAmount(nil, balance, decimals)
}
