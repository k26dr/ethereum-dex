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
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"dex/internal/amount"
	"dex/internal/prompt"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

type fillIn struct {
	contractAddress string
	walletAddress   string
	baseToken       common.Address
	quoteToken      common.Address
	orderID         *big.Int
	baseQuantityRaw string
	valueRaw        string
}

type fillOut struct {
	txHash     string
	minedBlock uint64
}

func newFillCommand(cfg *service.Service, ks *service.Keystore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fill",
		Short: "Fill an existing order",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFill(cmd, cfg, ks)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("wallet", "", "Wallet address to sign with")
	cmd.Flags().String("base", "", "Base token address or symbol")
	cmd.Flags().String("quote", "", "Quote token address or symbol")
	cmd.Flags().String("id", "", "Order id")
	cmd.Flags().String("base-qty", "", "Fill base quantity (decimal token units, e.g. 1.5)")
	cmd.Flags().String("value-wei", "", "Native value (decimal ETH units, optional, e.g. 0.1)")
	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	return cmd
}

func runFill(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) error {
	in, err := inputFill(cmd, cfg, ks)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	out, err := processFill(cmd, in, cfg, ks)
	if err != nil {
		return err
	}
	return outputFill(cmd, out)
}

func inputFill(cmd *cobra.Command, cfg *service.Service, ks *service.Keystore) (*fillIn, error) {
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
	baseQuantityRaw, _ := cmd.Flags().GetString("base-qty")
	baseQuantityRaw = strings.TrimSpace(baseQuantityRaw)
	valueRaw, err := orderReadAmountFlag(cmd, "value-wei", "Native value", false)
	if err != nil {
		return nil, err
	}

	return &fillIn{
		contractAddress: contractAddress,
		walletAddress:   walletAddress,
		baseToken:       baseToken,
		quoteToken:      quoteToken,
		orderID:         orderID,
		baseQuantityRaw: baseQuantityRaw,
		valueRaw:        valueRaw,
	}, nil
}

func processFill(cmd *cobra.Command, in *fillIn, cfg *service.Service, ks *service.Keystore) (*fillOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, orderRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()

	orderbookReadService, err := service.NewOrderBookService(rpcService, nil, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var walletForFillTx *service.Wallet
	orderValue, err := orderbookReadService.FetchOrder(ctx, in.baseToken, in.quoteToken, in.orderID)
	if err != nil {
		return nil, err
	}
	if fillOrderUnavailable(orderValue) {
		return nil, fmt.Errorf("order %s is already filled completely (or no longer exists)", in.orderID.String())
	}
	bankAddress, _, err := orderbookReadService.FetchBank(ctx, in.baseToken, in.quoteToken)
	if err != nil {
		return nil, err
	}

	baseMeta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, in.baseToken)
	if err != nil {
		return nil, fmt.Errorf("failed to read base token decimals: %w", err)
	}
	quoteMeta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, in.quoteToken)
	if err != nil {
		return nil, fmt.Errorf("failed to read quote token decimals: %w", err)
	}

	baseSymbol := fillTokenSymbolOrAddress(baseMeta.Symbol, in.baseToken)
	quoteSymbol := fillTokenSymbolOrAddress(quoteMeta.Symbol, in.quoteToken)

	var takerAction string
	var paymentToken common.Address
	var paymentSymbol string
	var paymentDecimals uint8
	switch orderValue.Side {
	case service.OrderSideSell:
		takerAction = "Buying " + baseSymbol + " with " + quoteSymbol
		paymentToken = in.quoteToken
		paymentSymbol = quoteSymbol
		paymentDecimals = quoteMeta.Decimals
	case service.OrderSideBuy:
		takerAction = "Selling " + baseSymbol + " for " + quoteSymbol
		paymentToken = in.baseToken
		paymentSymbol = baseSymbol
		paymentDecimals = baseMeta.Decimals
	}

	walletAddr := common.HexToAddress(in.walletAddress)
	paymentBalance, err := fillPaymentBalance(ctx, rpcService, cfg.Get().Network.ChainID, walletAddr, paymentToken)
	if err != nil {
		return nil, err
	}

	allowance := maxUint256()
	allowanceDisplay := "N/A (native token)"
	if paymentToken != (common.Address{}) {
		tokenReadService, err := service.NewTokenService(rpcService, nil, paymentToken, cfg.Get().Network.ChainID)
		if err != nil {
			return nil, err
		}
		allowance, err = tokenReadService.Allowance(ctx, walletAddr, common.HexToAddress(in.contractAddress))
		if err != nil {
			return nil, err
		}
		allowanceDisplay = fillFormatHuman(allowance, paymentDecimals) + " " + paymentSymbol
	}

	orderQuoteQuantity := fillDeriveQuoteQuantity(orderValue.BaseQuantity, orderValue.Price, quoteMeta.Decimals)

	maxByOrder := new(big.Int).Set(orderValue.BaseQuantity)
	maxByBalance := fillMaxBaseFromPayment(orderValue.Side, paymentBalance, orderValue.BaseQuantity, orderQuoteQuantity)
	maxFill := fillMinBigInt(maxByOrder, maxByBalance)
	if paymentToken != (common.Address{}) {
		maxByAllowance := fillMaxBaseFromPayment(orderValue.Side, allowance, orderValue.BaseQuantity, orderQuoteQuantity)
		maxFill = fillMinBigInt(maxFill, maxByAllowance)
	}
	requiredForFullFill := fillQuoteForBase(orderValue.BaseQuantity, orderValue.BaseQuantity, orderQuoteQuantity)
	if orderValue.Side == service.OrderSideBuy {
		requiredForFullFill = new(big.Int).Set(orderValue.BaseQuantity)
	}

	limitPrice := fillLimitPrice(orderValue.Price, quoteMeta.Decimals, orderValue.BaseQuantity, baseMeta.Decimals)

	fmt.Fprintf(cmd.OutOrStdout(), "Order ID: %s\n\n", in.orderID.String())
	fmt.Fprintln(cmd.OutOrStdout(), "Order details")
	fmt.Fprintf(cmd.OutOrStdout(), "  Market:      %s/%s\n", baseSymbol, quoteSymbol)
	fmt.Fprintf(cmd.OutOrStdout(), "  Maker side:  %s\n", fillTitle(fillMakerSide(orderValue.Side)))
	fmt.Fprintf(cmd.OutOrStdout(), "  Available:   %s %s\n", fillFormatHuman(orderValue.BaseQuantity, baseMeta.Decimals), baseSymbol)
	fmt.Fprintf(cmd.OutOrStdout(), "  Limit price: %s %s per %s\n", limitPrice, quoteSymbol, baseSymbol)
	fmt.Fprintf(cmd.OutOrStdout(), "  Total quote: %s %s\n\n", fillFormatHuman(orderQuoteQuantity, quoteMeta.Decimals), quoteSymbol)

	fmt.Fprintln(cmd.OutOrStdout(), "Your fill")
	fmt.Fprintf(cmd.OutOrStdout(), "  You are:     %s\n", takerAction)
	fmt.Fprintf(cmd.OutOrStdout(), "  Wallet:      %s\n", fillShortAddress(in.walletAddress))
	fmt.Fprintf(cmd.OutOrStdout(), "  Balance:     %s %s\n", fillFormatHuman(paymentBalance, paymentDecimals), paymentSymbol)
	fmt.Fprintf(cmd.OutOrStdout(), "  Allowance:   %s\n", allowanceDisplay)
	fmt.Fprintf(cmd.OutOrStdout(), "  Required:    %s %s (full remaining order)\n", fillFormatHuman(requiredForFullFill, paymentDecimals), paymentSymbol)
	fmt.Fprintf(cmd.OutOrStdout(), "  Max fill:    %s %s\n\n", fillFormatHuman(maxFill, baseMeta.Decimals), baseSymbol)

	if maxFill.Sign() <= 0 {
		if paymentToken == (common.Address{}) {
			return nil, fmt.Errorf("max fill is zero (check wallet balance)")
		}
		if paymentBalance.Sign() <= 0 {
			return nil, fmt.Errorf("max fill is zero (wallet %s balance is zero)", paymentSymbol)
		}

		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			ok, err := prompt.Confirm("Max fill is currently zero due to allowance. Set allowance now (this sends a separate approve transaction)")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("aborted")
			}
		}

		approveAmount, err := prompt.SelectAllowanceAmount(requiredForFullFill, paymentBalance, paymentDecimals)
		if err != nil {
			return nil, err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "\nReview allowance")
		fmt.Fprintf(cmd.OutOrStdout(), "  Token:       %s\n", paymentSymbol)
		fmt.Fprintf(cmd.OutOrStdout(), "  Amount:      %s %s\n", fillFormatHuman(approveAmount, paymentDecimals), paymentSymbol)
		fmt.Fprintf(cmd.OutOrStdout(), "  Spender:     %s\n\n", fillShortAddress(in.contractAddress))

		restore, err := fillPreparePasswordForWalletPrompt()
		if err != nil {
			return nil, err
		}
		defer restore()

		walletService, err := service.NewWallet(ks, in.walletAddress)
		if err != nil {
			return nil, err
		}
		walletForFillTx = walletService
		tokenWriteService, err := service.NewTokenService(rpcService, walletService, paymentToken, cfg.Get().Network.ChainID)
		if err != nil {
			return nil, err
		}

		approveTx, err := tokenWriteService.Approve(ctx, common.HexToAddress(in.contractAddress), approveAmount)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Approve tx: %s\n", approveTx.Hash().Hex())
		fmt.Fprintln(cmd.OutOrStdout(), "Waiting for approve transaction to be mined...")
		if _, err := service.WaitForTxSuccess(ctx, rpcService, approveTx.Hash(), service.DefaultTxWaitTimeout); err != nil {
			return nil, fmt.Errorf("approve transaction was not confirmed: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Approve confirmed.")

		tokenReadService, err := service.NewTokenService(rpcService, nil, paymentToken, cfg.Get().Network.ChainID)
		if err != nil {
			return nil, err
		}
		allowance, err = tokenReadService.Allowance(ctx, walletAddr, common.HexToAddress(in.contractAddress))
		if err != nil {
			return nil, err
		}
		allowanceDisplay = fillFormatHuman(allowance, paymentDecimals) + " " + paymentSymbol
		maxByAllowance := fillMaxBaseFromPayment(orderValue.Side, allowance, orderValue.BaseQuantity, orderQuoteQuantity)
		maxFill = fillMinBigInt(maxByOrder, fillMinBigInt(maxByBalance, maxByAllowance))
		fmt.Fprintf(cmd.OutOrStdout(), "Updated allowance: %s\n", allowanceDisplay)
		fmt.Fprintf(cmd.OutOrStdout(), "Updated max fill:  %s %s\n\n", fillFormatHuman(maxFill, baseMeta.Decimals), baseSymbol)
		if maxFill.Sign() <= 0 {
			return nil, fmt.Errorf("max fill is still zero after approve")
		}
	}

	baseQuantityRaw := strings.TrimSpace(in.baseQuantityRaw)
	if baseQuantityRaw == "" {
		defaultAmount := fillFormatHuman(maxFill, baseMeta.Decimals)
		label := "Amount to buy"
		if orderValue.Side == service.OrderSideBuy {
			label = "Amount to sell"
		}
		value, err := fillPromptWithDefault(fmt.Sprintf("%s in %s [default: %s]", label, baseSymbol, defaultAmount), defaultAmount)
		if err != nil {
			return nil, err
		}
		baseQuantityRaw = value
	}

	baseQuantity, err := amount.ParseUnits(baseQuantityRaw, baseMeta.Decimals)
	if err != nil {
		return nil, fmt.Errorf("invalid fill base quantity %q: %w", baseQuantityRaw, err)
	}
	if baseQuantity.Sign() <= 0 {
		return nil, fmt.Errorf("fill amount must be greater than zero")
	}
	if baseQuantity.Cmp(orderValue.BaseQuantity) > 0 {
		return nil, fmt.Errorf("fill amount exceeds available order size")
	}
	if baseQuantity.Cmp(maxFill) > 0 {
		return nil, fmt.Errorf(
			"fill amount exceeds your max fill (%s %s)",
			fillFormatHuman(maxFill, baseMeta.Decimals),
			baseSymbol,
		)
	}

	// Final pre-send refresh: order can change between prompt and submission.
	latestOrder, err := orderbookReadService.FetchOrder(ctx, in.baseToken, in.quoteToken, in.orderID)
	if err != nil {
		return nil, err
	}
	if latestOrder.User == (common.Address{}) || latestOrder.BaseQuantity == nil || latestOrder.BaseQuantity.Sign() <= 0 {
		return nil, fmt.Errorf("order %s is already filled or no longer available", in.orderID.String())
	}
	if latestOrder.Side != orderValue.Side {
		return nil, fmt.Errorf("order %s changed unexpectedly; please retry", in.orderID.String())
	}
	if baseQuantity.Cmp(latestOrder.BaseQuantity) > 0 {
		return nil, fmt.Errorf(
			"order %s changed: available now %s %s, requested %s %s",
			in.orderID.String(),
			fillFormatHuman(latestOrder.BaseQuantity, baseMeta.Decimals),
			baseSymbol,
			fillFormatHuman(baseQuantity, baseMeta.Decimals),
			baseSymbol,
		)
	}

	latestQuoteQuantity := fillDeriveQuoteQuantity(latestOrder.BaseQuantity, latestOrder.Price, quoteMeta.Decimals)
	fillQuote := fillQuoteForBase(baseQuantity, latestOrder.BaseQuantity, latestQuoteQuantity)
	requiredPayment := fillQuote
	if latestOrder.Side == service.OrderSideBuy {
		requiredPayment = new(big.Int).Set(baseQuantity)
	}
	latestLimitPrice := fillLimitPrice(latestOrder.Price, quoteMeta.Decimals, latestOrder.BaseQuantity, baseMeta.Decimals)

	currentBalance, err := fillPaymentBalance(ctx, rpcService, cfg.Get().Network.ChainID, walletAddr, paymentToken)
	if err != nil {
		return nil, err
	}
	if paymentToken != (common.Address{}) {
		tokenReadService, err := service.NewTokenService(rpcService, nil, paymentToken, cfg.Get().Network.ChainID)
		if err != nil {
			return nil, err
		}
		currentAllowance, err := tokenReadService.Allowance(ctx, walletAddr, common.HexToAddress(in.contractAddress))
		if err != nil {
			return nil, err
		}
		if currentAllowance.Cmp(requiredPayment) < 0 {
			return nil, fmt.Errorf(
				"allowance changed and is now too low: need %s %s, have %s %s",
				fillFormatHuman(requiredPayment, paymentDecimals),
				paymentSymbol,
				fillFormatHuman(currentAllowance, paymentDecimals),
				paymentSymbol,
			)
		}
	}
	if currentBalance.Cmp(requiredPayment) < 0 {
		return nil, fmt.Errorf(
			"balance changed and is now too low: need %s %s, have %s %s",
			fillFormatHuman(requiredPayment, paymentDecimals),
			paymentSymbol,
			fillFormatHuman(currentBalance, paymentDecimals),
			paymentSymbol,
		)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "\nReview fill")
	if latestOrder.Side == service.OrderSideSell {
		fmt.Fprintf(cmd.OutOrStdout(), "  You buy:     %s %s\n", fillFormatHuman(baseQuantity, baseMeta.Decimals), baseSymbol)
		fmt.Fprintf(cmd.OutOrStdout(), "  You pay:     %s %s\n", fillFormatHuman(requiredPayment, paymentDecimals), paymentSymbol)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "  You sell:    %s %s\n", fillFormatHuman(baseQuantity, baseMeta.Decimals), baseSymbol)
		fmt.Fprintf(cmd.OutOrStdout(), "  You receive: %s %s\n", fillFormatHuman(fillQuote, quoteMeta.Decimals), quoteSymbol)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "  Price:       %s %s per %s\n", latestLimitPrice, quoteSymbol, baseSymbol)
	fmt.Fprintf(cmd.OutOrStdout(), "  Order ID:    %s\n", in.orderID.String())
	fmt.Fprintf(cmd.OutOrStdout(), "  Market bank: %s\n\n", bankAddress.Hex())

	value := big.NewInt(0)
	trimmedValueRaw := strings.TrimSpace(in.valueRaw)
	if paymentToken == (common.Address{}) {
		if trimmedValueRaw == "" {
			value = requiredPayment
		} else {
			value, err = amount.ParseUnits(trimmedValueRaw, 18)
			if err != nil {
				return nil, fmt.Errorf("invalid native value %q: %w", trimmedValueRaw, err)
			}
			if value.Cmp(requiredPayment) != 0 {
				return nil, fmt.Errorf("native value must equal required payment %s", fillFormatHuman(requiredPayment, 18))
			}
		}
	} else if trimmedValueRaw != "" {
		value, err = amount.ParseUnits(trimmedValueRaw, 18)
		if err != nil {
			return nil, fmt.Errorf("invalid native value %q: %w", trimmedValueRaw, err)
		}
		if value.Sign() != 0 {
			return nil, fmt.Errorf("native value should not be set for token-funded fills")
		}
	}

	var walletService *service.Wallet
	if walletForFillTx != nil {
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			ok, err := prompt.Confirm("Wallet already unlocked from allowance step. Proceed with fill transaction")
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("aborted")
			}
		}
		walletService = walletForFillTx
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

	orderbookService, err := service.NewOrderBookService(rpcService, walletService, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	tx, err := orderbookService.FillOrder(ctx, in.orderID, in.baseToken, in.quoteToken, baseQuantity, value)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Fill order tx submitted: %s\n", tx.Hash().Hex())
	fmt.Fprintln(cmd.OutOrStdout(), "Waiting for fill transaction to be mined...")
	receipt, err := service.WaitForTxSuccess(ctx, rpcService, tx.Hash(), service.DefaultTxWaitTimeout)
	if err != nil {
		return nil, err
	}
	return &fillOut{txHash: tx.Hash().Hex(), minedBlock: receipt.BlockNumber.Uint64()}, nil
}

func outputFill(cmd *cobra.Command, out *fillOut) error {
	fmt.Fprintf(cmd.OutOrStdout(), "\nFill order tx: %s\n", out.txHash)
	fmt.Fprintf(cmd.OutOrStdout(), "Fill order mined in block: %d\n", out.minedBlock)
	return nil
}

func fillTokenSymbolOrAddress(symbol string, token common.Address) string {
	trimmed := strings.TrimSpace(symbol)
	if trimmed != "" {
		return trimmed
	}
	if token == (common.Address{}) {
		return "ETH"
	}
	return token.Hex()
}

func fillPaymentBalance(ctx context.Context, rpc *service.RPC, chainID int64, wallet common.Address, token common.Address) (*big.Int, error) {
	if token == (common.Address{}) {
		return rpc.Client().BalanceAt(ctx, wallet, nil)
	}
	tokenService, err := service.NewTokenService(rpc, nil, token, chainID)
	if err != nil {
		return nil, err
	}
	return tokenService.BalanceOf(ctx, wallet)
}

func fillMakerSide(side uint8) string {
	if side == service.OrderSideSell {
		return "sell"
	}
	if side == service.OrderSideBuy {
		return "buy"
	}
	return "unknown"
}

func fillTitle(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func fillQuoteForBase(baseQuantity *big.Int, orderBase *big.Int, orderQuote *big.Int) *big.Int {
	if orderBase.Sign() == 0 {
		return big.NewInt(0)
	}
	q := new(big.Int).Mul(baseQuantity, orderQuote)
	return q.Div(q, orderBase)
}

func fillMaxBaseFromPayment(side uint8, paymentAmount *big.Int, orderBase *big.Int, orderQuote *big.Int) *big.Int {
	if paymentAmount == nil || paymentAmount.Sign() <= 0 {
		return big.NewInt(0)
	}
	if side == service.OrderSideBuy {
		return new(big.Int).Set(paymentAmount)
	}
	if orderQuote == nil || orderQuote.Sign() == 0 {
		return big.NewInt(0)
	}
	n := new(big.Int).Mul(paymentAmount, orderBase)
	return n.Div(n, orderQuote)
}

func fillMinBigInt(a, b *big.Int) *big.Int {
	if a == nil {
		return new(big.Int).Set(b)
	}
	if b == nil {
		return new(big.Int).Set(a)
	}
	if a.Cmp(b) <= 0 {
		return new(big.Int).Set(a)
	}
	return new(big.Int).Set(b)
}

func fillLimitPrice(orderQuote *big.Int, quoteDecimals uint8, orderBase *big.Int, baseDecimals uint8) string {
	if orderBase == nil || orderBase.Sign() == 0 || orderQuote == nil {
		return "0.0"
	}
	num := new(big.Int).Set(orderQuote)
	num.Mul(num, pow10Big(baseDecimals))
	den := new(big.Int).Set(orderBase)
	den.Mul(den, pow10Big(quoteDecimals))
	r := new(big.Rat).SetFrac(num, den)
	s := r.FloatString(8)
	return fillTrimDecimal(s)
}

func pow10Big(n uint8) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func fillTrimDecimal(s string) string {
	if !strings.Contains(s, ".") {
		return s + ".0"
	}
	s = strings.TrimRight(s, "0")
	if strings.HasSuffix(s, ".") {
		return s + "0"
	}
	return s
}

func fillFormatHuman(raw *big.Int, decimals uint8) string {
	return fillWithThousands(amount.FormatUnits(raw, decimals))
}

func fillWithThousands(s string) string {
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = strings.TrimPrefix(s, "-")
	}
	parts := strings.SplitN(s, ".", 2)
	whole := parts[0]
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}

	var b strings.Builder
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteRune(',')
		}
		b.WriteRune(r)
	}
	if frac != "" {
		return sign + b.String() + "." + frac
	}
	return sign + b.String()
}

func fillShortAddress(addr string) string {
	addr = strings.TrimSpace(addr)
	if len(addr) <= 12 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-5:]
}

func fillPromptWithDefault(label string, defaultValue string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s: ", label)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("aborted")
	}
	value := strings.TrimSpace(scanner.Text())
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func fillPreparePasswordForWalletPrompt() (func(), error) {
	if strings.TrimSpace(os.Getenv(service.WalletPasswordEnvVar)) != "" {
		return func() {}, nil
	}

	pw, err := prompt.Password("Enter Password to decrypt your keystore and proceed")
	if err != nil {
		return nil, err
	}
	prev, hadPrev := os.LookupEnv(service.WalletPasswordEnvVar)
	_ = os.Setenv(service.WalletPasswordEnvVar, pw)
	return func() {
		if hadPrev {
			_ = os.Setenv(service.WalletPasswordEnvVar, prev)
		} else {
			_ = os.Unsetenv(service.WalletPasswordEnvVar)
		}
	}, nil
}

func maxUint256() *big.Int {
	max := new(big.Int).Lsh(big.NewInt(1), 256)
	return max.Sub(max, big.NewInt(1))
}

func fillDeriveQuoteQuantity(baseQuantity *big.Int, price *big.Int, quoteDecimals uint8) *big.Int {
	if baseQuantity == nil || price == nil {
		return big.NewInt(0)
	}
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(quoteDecimals)), nil)
	q := new(big.Int).Mul(baseQuantity, price)
	return q.Div(q, exp)
}

func fillOrderUnavailable(orderValue *service.Order) bool {
	if orderValue == nil {
		return true
	}
	if orderValue.User == (common.Address{}) {
		return true
	}
	if orderValue.BaseQuantity == nil || orderValue.BaseQuantity.Sign() <= 0 {
		return true
	}
	return false
}

