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
	"time"

	"dex/internal/amount"
	"dex/internal/prompt"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const orderRPCConnectTimeout = 15 * time.Second

type placeIn struct {
	contractAddress string
	walletAddress   string
	baseToken       common.Address
	quoteToken      common.Address
	side            uint8
	baseQuantityRaw string
	priceRaw        string
	valueRaw        string
}

type placeOut struct {
	bankAddress       string
	createdMarket     bool
	createMarketTx    string
	createMarketBlock uint64
	allowanceAdjusted bool
	allowanceToken    string
	allowanceAmount   string
	allowanceDisplay  string
	approveTx         string
	approveBlock      uint64
	placeOrderTx      string
	placeOrderBlock   uint64
}

func newPlaceCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place",
		Short: "Place a new order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlace(cmd, cfg, ks)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().String("base", "", "Base token address or symbol")
	cmd.Flags().String("quote", "", "Quote token address or symbol")
	cmd.Flags().String("side", "", "Order side: buy or sell")
	cmd.Flags().String("base-qty", "", "Base token quantity (decimal token units, e.g. 1.5)")
	cmd.Flags().String("quote-qty", "", "Quote token quantity (decimal token units, e.g. 2.0)")
	cmd.Flags().String("value-wei", "", "Native value (decimal ETH units, optional, e.g. 0.1)")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	return cmd
}

func runPlace(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	in, err := inputPlace(cmd, cfg, ks)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	out, err := processPlace(cmd, in, cfg, ks)
	if err != nil {
		return err
	}
	return outputPlace(cmd, out)
}

func inputPlace(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) (*placeIn, error) {
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
	if baseToken == quoteToken {
		return nil, fmt.Errorf("base and quote token must be different")
	}
	side, err := orderReadSide(cmd)
	if err != nil {
		return nil, err
	}
	baseSymbol := orderTokenPromptLabel(cfg, baseToken)
	quoteSymbol := orderTokenPromptLabel(cfg, quoteToken)
	baseAmountPrompt := fmt.Sprintf("Amount to buy (%s)", baseSymbol)
	if side == service.OrderSideSell {
		baseAmountPrompt = fmt.Sprintf("Amount to sell (%s)", baseSymbol)
	}
	limitPricePrompt := fmt.Sprintf("Limit price (%s per %s)", quoteSymbol, baseSymbol)

	baseQuantityRaw, err := orderReadAmountFlag(cmd, "base-qty", baseAmountPrompt, true)
	if err != nil {
		return nil, err
	}
	priceRaw, err := orderReadAmountFlag(cmd, "quote-qty", limitPricePrompt, true)
	if err != nil {
		return nil, err
	}
	valueRaw, err := orderReadAmountFlag(cmd, "value-wei", "Native value", false)
	if err != nil {
		return nil, err
	}

	return &placeIn{
		contractAddress: contractAddress,
		walletAddress:   walletAddress,
		baseToken:       baseToken,
		quoteToken:      quoteToken,
		side:            side,
		baseQuantityRaw: baseQuantityRaw,
		priceRaw:        priceRaw,
		valueRaw:        valueRaw,
	}, nil
}

func processPlace(cmd *cobra.Command, in *placeIn, cfg *service.Service, ks *service.Keystore) (*placeOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, orderRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()

	orderbookReadService, err := service.NewOrderBookService(rpcService, nil, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}
	var walletForPlaceTx *service.Wallet

	ctx := context.Background()
	baseMeta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, in.baseToken)
	if err != nil {
		return nil, fmt.Errorf("failed to read base token decimals: %w", err)
	}
	quoteMeta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, in.quoteToken)
	if err != nil {
		return nil, fmt.Errorf("failed to read quote token decimals: %w", err)
	}
	baseDecimals := baseMeta.Decimals
	quoteDecimals := quoteMeta.Decimals

	baseQuantity, err := amount.ParseUnits(in.baseQuantityRaw, baseDecimals)
	if err != nil {
		return nil, fmt.Errorf("invalid base quantity %q: %w", in.baseQuantityRaw, err)
	}
	price, err := amount.ParseUnits(in.priceRaw, quoteDecimals)
	if err != nil {
		return nil, fmt.Errorf("invalid price %q: %w", in.priceRaw, err)
	}
	// quoteQuantity = baseQuantity * price / 10**quoteDecimals (mirrors contract logic, used for allowance)
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(quoteDecimals)), nil)
	quoteQuantity := new(big.Int).Div(new(big.Int).Mul(baseQuantity, price), exp)

	value := big.NewInt(0)
	if strings.TrimSpace(in.valueRaw) != "" {
		value, err = amount.ParseUnits(in.valueRaw, 18)
		if err != nil {
			return nil, fmt.Errorf("invalid native value %q: %w", in.valueRaw, err)
		}
	}

	bankAddress, deployed, err := orderbookReadService.FetchBank(ctx, in.baseToken, in.quoteToken)
	if err != nil {
		return nil, err
	}

	out := &placeOut{
		bankAddress: bankAddress.Hex(),
	}

	allowanceToken, requiredAllowance, needsAllowance := placeRequiredAllowance(in.side, in.baseToken, in.quoteToken, baseQuantity, quoteQuantity, value)
	if needsAllowance {
		allowanceSymbol := strings.TrimSpace(baseMeta.Symbol)
		if in.side == service.OrderSideBuy {
			allowanceSymbol = strings.TrimSpace(quoteMeta.Symbol)
		}
		allowanceRef := service.FormatTokenRef(allowanceSymbol, allowanceToken.Hex())

		tokenService, err := service.NewTokenService(rpcService, nil, allowanceToken, cfg.Get().Network.ChainID)
		if err != nil {
			return nil, err
		}

		owner := common.HexToAddress(in.walletAddress)
		spender := common.HexToAddress(in.contractAddress)
		currentAllowance, err := tokenService.Allowance(ctx, owner, spender)
		if err != nil {
			return nil, err
		}
		allowanceDecimals := baseDecimals
		if in.side == service.OrderSideBuy {
			allowanceDecimals = quoteDecimals
		}

		if currentAllowance.Cmp(requiredAllowance) < 0 {
			fmt.Fprintf(
				cmd.ErrOrStderr(),
				"Allowance too low for token %s (current: %s (raw: %s), required: %s (raw: %s))\n",
				allowanceRef,
				amount.FormatUnits(currentAllowance, allowanceDecimals),
				currentAllowance.String(),
				amount.FormatUnits(requiredAllowance, allowanceDecimals),
				requiredAllowance.String(),
			)
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				ok, err := prompt.Confirm("Set allowance now (this sends a separate approve transaction before placing the order)")
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, fmt.Errorf("aborted")
				}
			}

			balance, err := tokenService.BalanceOf(ctx, owner)
			if err != nil {
				return nil, err
			}

			var approveAmount *big.Int
			for {
				approveAmount, err = prompt.SelectAllowanceAmount(requiredAllowance, balance, allowanceDecimals)
				if err != nil {
					return nil, err
				}
				if approveAmount.Cmp(requiredAllowance) < 0 {
					fmt.Fprintf(
						cmd.ErrOrStderr(),
						"Selected allowance (%s (raw: %s)) is below required order amount (%s (raw: %s)).\n",
						amount.FormatUnits(approveAmount, allowanceDecimals),
						approveAmount.String(),
						amount.FormatUnits(requiredAllowance, allowanceDecimals),
						requiredAllowance.String(),
					)
					continue
				}
				break
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Review allowance")
			fmt.Fprintf(cmd.OutOrStdout(), "  Wallet:  %s\n", in.walletAddress)
			fmt.Fprintf(cmd.OutOrStdout(), "  Token:   %s\n", allowanceRef)
			fmt.Fprintf(cmd.OutOrStdout(), "  Amount:  %s (raw: %s)\n", amount.FormatUnits(approveAmount, allowanceDecimals), approveAmount.String())
			ok, err := prompt.Confirm("Proceed and send approve transaction")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("aborted")
			}

			restore, err := fillPreparePasswordForWalletPrompt()
			if err != nil {
				return nil, err
			}
			defer restore()
			walletService, err := service.NewWallet(ks, in.walletAddress)
			if err != nil {
				return nil, err
			}
			walletForPlaceTx = walletService
			tokenWriteService, err := service.NewTokenService(rpcService, walletService, allowanceToken, cfg.Get().Network.ChainID)
			if err != nil {
				return nil, err
			}

			approveTx, err := tokenWriteService.Approve(ctx, spender, approveAmount)
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Approve tx submitted: %s\n", approveTx.Hash().Hex())
			fmt.Fprintln(cmd.OutOrStdout(), "Waiting for approve transaction to be mined...")
			approveReceipt, err := service.WaitForTxSuccess(ctx, rpcService, approveTx.Hash(), service.DefaultTxWaitTimeout)
			if err != nil {
				return nil, err
			}

			out.allowanceAdjusted = true
			out.allowanceToken = allowanceRef
			out.allowanceAmount = approveAmount.String()
			out.allowanceDisplay = amount.FormatUnits(approveAmount, allowanceDecimals)
			out.approveTx = approveTx.Hash().Hex()
			out.approveBlock = approveReceipt.BlockNumber.Uint64()
		}
	}

	if !deployed {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			ok, err := prompt.Confirm("Market is not deployed for this pair. Deploy market now")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("aborted")
			}
		}

		var walletService *service.Wallet
		if walletForPlaceTx != nil {
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes {
				ok, err := prompt.Confirm("Wallet already unlocked from previous step. Proceed with create market transaction")
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, fmt.Errorf("aborted")
				}
			}
			walletService = walletForPlaceTx
		} else {
			restore, err := fillPreparePasswordForWalletPrompt()
			if err != nil {
				return nil, err
			}
			defer restore()
			walletService, err = service.NewWallet(ks, in.walletAddress)
			if err != nil {
				return nil, err
			}
			walletForPlaceTx = walletService
		}
		orderbookWriteService, err := service.NewOrderBookService(rpcService, walletService, in.contractAddress, cfg.Get().Network.ChainID)
		if err != nil {
			return nil, err
		}
		createTx, err := orderbookWriteService.CreateMarket(ctx, in.baseToken, in.quoteToken)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Create market tx submitted: %s\n", createTx.Hash().Hex())
		fmt.Fprintln(cmd.OutOrStdout(), "Waiting for create market transaction to be mined...")
		createReceipt, err := service.WaitForTxSuccess(ctx, rpcService, createTx.Hash(), service.DefaultTxWaitTimeout)
		if err != nil {
			return nil, err
		}
		out.createdMarket = true
		out.createMarketTx = createTx.Hash().Hex()
		out.createMarketBlock = createReceipt.BlockNumber.Uint64()
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Review place order")
	fmt.Fprintf(cmd.OutOrStdout(), "  Wallet: %s\n", in.walletAddress)
	fmt.Fprintf(cmd.OutOrStdout(), "  Side:   %s\n", strings.ToUpper(map[uint8]string{service.OrderSideBuy: "buy", service.OrderSideSell: "sell"}[in.side]))
	fmt.Fprintf(cmd.OutOrStdout(), "  Base:   %s\n", in.baseToken.Hex())
	fmt.Fprintf(cmd.OutOrStdout(), "  Quote:  %s\n", in.quoteToken.Hex())
	fmt.Fprintf(cmd.OutOrStdout(), "  Amount: %s (raw: %s)\n", amount.FormatUnits(baseQuantity, baseDecimals), baseQuantity.String())
	fmt.Fprintf(cmd.OutOrStdout(), "  Price:  %s (raw: %s)\n", amount.FormatUnits(price, quoteDecimals), price.String())

	var walletService *service.Wallet
	if walletForPlaceTx != nil {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			ok, err := prompt.Confirm("Wallet already unlocked from previous step. Proceed with place order transaction")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("aborted")
			}
		}
		walletService = walletForPlaceTx
	} else {
		restore, err := fillPreparePasswordForWalletPrompt()
		if err != nil {
			return nil, err
		}
		defer restore()
		walletService, err = service.NewWallet(ks, in.walletAddress)
		if err != nil {
			return nil, err
		}
	}
	orderbookWriteService, err := service.NewOrderBookService(rpcService, walletService, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	placeTx, err := orderbookWriteService.PlaceOrder(ctx, in.baseToken, in.quoteToken, in.side, baseQuantity, price, value)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Place order tx submitted: %s\n", placeTx.Hash().Hex())
	fmt.Fprintln(cmd.OutOrStdout(), "Waiting for place order transaction to be mined...")
	placeReceipt, err := service.WaitForTxSuccess(ctx, rpcService, placeTx.Hash(), service.DefaultTxWaitTimeout)
	if err != nil {
		return nil, err
	}
	out.placeOrderTx = placeTx.Hash().Hex()
	out.placeOrderBlock = placeReceipt.BlockNumber.Uint64()

	return out, nil
}

func outputPlace(cmd *cobra.Command, out *placeOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "Market bank: %s\n", out.bankAddress)
	if out.createdMarket {
		fmt.Fprintf(cmd.OutOrStdout(), "Created market tx: %s\n", out.createMarketTx)
		fmt.Fprintf(cmd.OutOrStdout(), "Created market mined in block: %d\n", out.createMarketBlock)
	}
	if out.allowanceAdjusted {
		fmt.Fprintf(cmd.OutOrStdout(), "Allowance token: %s\n", out.allowanceToken)
		fmt.Fprintf(cmd.OutOrStdout(), "Allowance amount: %s (raw: %s)\n", out.allowanceDisplay, out.allowanceAmount)
		fmt.Fprintf(cmd.OutOrStdout(), "Approve tx: %s\n", out.approveTx)
		fmt.Fprintf(cmd.OutOrStdout(), "Approve mined in block: %d\n", out.approveBlock)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Place order tx: %s\n", out.placeOrderTx)
	fmt.Fprintf(cmd.OutOrStdout(), "Place order mined in block: %d\n", out.placeOrderBlock)
	return nil
}

func placeRequiredAllowance(side uint8, baseToken common.Address, quoteToken common.Address, baseQuantity *big.Int, quoteQuantity *big.Int, value *big.Int) (common.Address, *big.Int, bool) {
	if side == service.OrderSideSell && value != nil && value.Sign() == 0 && baseToken != (common.Address{}) {
		return baseToken, baseQuantity, true
	}
	if side == service.OrderSideBuy && value != nil && value.Sign() == 0 && quoteToken != (common.Address{}) {
		return quoteToken, quoteQuantity, true
	}
	return common.Address{}, nil, false
}
