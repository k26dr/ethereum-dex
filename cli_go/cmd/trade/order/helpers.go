// Copyright © 2026 0xTrooper (on Github)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package order

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func orderReadContractAddress(cmd *cobra.Command, cfg *service.Config) (string, error) {
	contractAddress, _ := cmd.Flags().GetString("contract")
	contractAddress = strings.TrimSpace(contractAddress)
	if contractAddress == "" && cfg != nil {
		configAddress := strings.TrimSpace(cfg.Contract.Address)
		if common.IsHexAddress(configAddress) && common.HexToAddress(configAddress) != (common.Address{}) {
			contractAddress = configAddress
		}
	}
	for contractAddress == "" {
		next, err := orderAsk("OrderBook contract address")
		if err != nil {
			return "", err
		}
		contractAddress = strings.TrimSpace(next)
	}
	if !common.IsHexAddress(contractAddress) {
		return "", fmt.Errorf("invalid contract address %q", contractAddress)
	}
	if common.HexToAddress(contractAddress) == (common.Address{}) {
		return "", fmt.Errorf("contract address cannot be zero address")
	}
	return contractAddress, nil
}

func orderReadAddressFlag(cmd *cobra.Command, flagName string, promptLabel string) (common.Address, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			next, err := orderAsk(promptLabel)
			if err != nil {
				return common.Address{}, err
			}
			value = strings.TrimSpace(next)
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return common.Address{}, err
			}
		}
		if !common.IsHexAddress(value) {
			fmt.Fprintf(os.Stderr, "Invalid address %q.\n", value)
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return common.Address{}, err
			}
			continue
		}
		return common.HexToAddress(value), nil
	}
}

func orderReadTokenAddressFlag(cmd *cobra.Command, cfg *service.Service, flagName string, promptLabel string) (common.Address, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			next, err := orderAsk(promptLabel)
			if err != nil {
				return common.Address{}, err
			}
			value = strings.TrimSpace(next)
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return common.Address{}, err
			}
		}

		if common.IsHexAddress(value) {
			return common.HexToAddress(value), nil
		}

		address, matches := orderResolveTokenSymbol(cfg, value)
		if matches == 1 {
			if err := cmd.Flags().Set(flagName, address.Hex()); err != nil {
				return common.Address{}, err
			}
			return address, nil
		}
		if matches == 0 {
			fmt.Fprintf(os.Stderr, "Unknown token symbol %q. Please enter a token address.\n", value)
		} else {
			fmt.Fprintf(os.Stderr, "Token symbol %q matches multiple tokens. Please enter a token address.\n", value)
		}

		if err := cmd.Flags().Set(flagName, ""); err != nil {
			return common.Address{}, err
		}
	}
}

func orderResolveTokenSymbol(cfg *service.Service, symbol string) (common.Address, int) {
	if cfg == nil || cfg.Get() == nil {
		return common.Address{}, 0
	}
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return common.Address{}, 0
	}

	tokens := service.MergeKnownAndConfiguredTokens(cfg.Get().Network.ChainID, cfg.Get().Tokens)
	match := common.Address{}
	count := 0
	for _, token := range tokens {
		if !strings.EqualFold(strings.TrimSpace(token.Symbol), symbol) {
			continue
		}
		if !common.IsHexAddress(token.Address) {
			continue
		}
		count++
		if count == 1 {
			match = common.HexToAddress(token.Address)
		}
	}
	return match, count
}

func orderTokenPromptLabel(cfg *service.Service, token common.Address) string {
	if cfg == nil || cfg.Get() == nil {
		return token.Hex()
	}
	for _, t := range service.MergeKnownAndConfiguredTokens(cfg.Get().Network.ChainID, cfg.Get().Tokens) {
		if !common.IsHexAddress(t.Address) {
			continue
		}
		if common.HexToAddress(t.Address) != token {
			continue
		}
		symbol := strings.TrimSpace(t.Symbol)
		if symbol != "" {
			return symbol
		}
		break
	}
	return token.Hex()
}

func orderReadBigIntFlag(cmd *cobra.Command, flagName string, promptLabel string, mustBePositive bool) (*big.Int, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			if !mustBePositive {
				return big.NewInt(0), nil
			}
			next, err := orderAsk(promptLabel)
			if err != nil {
				return nil, err
			}
			value = strings.TrimSpace(next)
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return nil, err
			}
		}

		n, ok := new(big.Int).SetString(value, 10)
		if !ok {
			fmt.Fprintf(os.Stderr, "Invalid integer value %q.\n", value)
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return nil, err
			}
			continue
		}
		if mustBePositive && n.Sign() <= 0 {
			fmt.Fprintln(os.Stderr, "Value must be greater than zero.")
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return nil, err
			}
			continue
		}
		if !mustBePositive && n.Sign() < 0 {
			fmt.Fprintln(os.Stderr, "Value cannot be negative.")
			if err := cmd.Flags().Set(flagName, ""); err != nil {
				return nil, err
			}
			continue
		}
		return n, nil
	}
}

func orderReadAmountFlag(cmd *cobra.Command, flagName string, promptLabel string, required bool) (string, error) {
	for {
		value, _ := cmd.Flags().GetString(flagName)
		value = strings.TrimSpace(value)
		if value == "" {
			if !required {
				return "", nil
			}
			next, err := orderAsk(promptLabel)
			if err != nil {
				return "", err
			}
			value = strings.TrimSpace(next)
			if err := cmd.Flags().Set(flagName, value); err != nil {
				return "", err
			}
			if value == "" {
				continue
			}
		}
		return value, nil
	}
}

func orderReadSide(cmd *cobra.Command) (uint8, error) {
	for {
		value, _ := cmd.Flags().GetString("side")
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			next, err := orderAsk("Side [buy/sell]")
			if err != nil {
				return 0, err
			}
			value = strings.ToLower(strings.TrimSpace(next))
			if err := cmd.Flags().Set("side", value); err != nil {
				return 0, err
			}
		}
		switch value {
		case "buy", "0":
			return service.OrderSideBuy, nil
		case "sell", "1":
			return service.OrderSideSell, nil
		default:
			fmt.Fprintln(os.Stderr, "Invalid side. Please use buy or sell.")
			if err := cmd.Flags().Set("side", ""); err != nil {
				return 0, err
			}
		}
	}
}

func orderAsk(label string) (string, error) {
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

func orderSelectWalletAddress(ks *service.Keystore, requestedAddress string) (string, error) {
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
