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

	"dex/internal/amount"
	"dex/service"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

type getIn struct {
	contractAddress string
	baseToken       common.Address
	quoteToken      common.Address
	orderID         *big.Int
}

type getOut struct {
	status              string
	active              bool
	user                string
	baseToken           string
	quoteToken          string
	baseSymbol          string
	quoteSymbol         string
	side                string
	baseQuantityRaw     string
	baseQuantityDisplay string
	priceRaw            string
	priceDisplay        string
}

func newGetCommand(cfg *service.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a specific order",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				currentID, _ := cmd.Flags().GetString("id")
				if strings.TrimSpace(currentID) == "" {
					if err := cmd.Flags().Set("id", strings.TrimSpace(args[0])); err != nil {
						return err
					}
				}
			}
			return runGet(cmd, cfg)
		},
	}

	cmd.Flags().String("contract", "", "OrderBook contract address (defaults to config.contract.address)")
	cmd.Flags().String("base", "", "Base token address or symbol")
	cmd.Flags().String("quote", "", "Quote token address or symbol")
	cmd.Flags().String("id", "", "Order id")
	return cmd
}

func runGet(cmd *cobra.Command, cfg *service.Service) error {
	in, err := inputGet(cmd, cfg)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true
	out, err := processGet(in, cfg)
	if err != nil {
		return err
	}
	return outputGet(cmd, out)
}

func inputGet(cmd *cobra.Command, cfg *service.Service) (*getIn, error) {
	contractAddress, err := orderReadContractAddress(cmd, cfg.Get())
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
	return &getIn{contractAddress: contractAddress, baseToken: baseToken, quoteToken: quoteToken, orderID: orderID}, nil
}

func processGet(in *getIn, cfg *service.Service) (*getOut, error) {
	rpcService, err := service.NewRPC(cfg.Get().Network, orderRPCConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer rpcService.Close()

	orderbookService, err := service.NewOrderBookService(rpcService, nil, in.contractAddress, cfg.Get().Network.ChainID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	orderValue, err := orderbookService.FetchOrder(ctx, in.baseToken, in.quoteToken, in.orderID)
	if err != nil {
		return nil, err
	}

	side := "unknown"
	active := orderValue.User != (common.Address{}) && orderValue.BaseQuantity != nil && orderValue.BaseQuantity.Sign() > 0
	status := "active"
	if !active {
		status = "inactive (order not found, canceled, or fully filled)"
		side = "n/a"
	} else if orderValue.Side == service.OrderSideBuy {
		side = "buy"
	} else if orderValue.Side == service.OrderSideSell {
		side = "sell"
	}

	baseDecimals := uint8(18)
	baseSymbol := ""
	if meta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, in.baseToken); err == nil {
		if meta.Decimals > 0 {
			baseDecimals = meta.Decimals
		}
		baseSymbol = strings.TrimSpace(meta.Symbol)
	}
	quoteDecimals := uint8(18)
	quoteSymbol := ""
	if meta, err := cfg.ResolveTokenMetadata(ctx, rpcService, cfg.Get().Network.ChainID, in.quoteToken); err == nil {
		if meta.Decimals > 0 {
			quoteDecimals = meta.Decimals
		}
		quoteSymbol = strings.TrimSpace(meta.Symbol)
	}

	return &getOut{
		status:              status,
		active:              active,
		user:                orderValue.User.Hex(),
		baseToken:           service.FormatTokenRef(baseSymbol, in.baseToken.Hex()),
		quoteToken:          service.FormatTokenRef(quoteSymbol, in.quoteToken.Hex()),
		baseSymbol:          baseSymbol,
		quoteSymbol:         quoteSymbol,
		side:                side,
		baseQuantityRaw:     orderValue.BaseQuantity.String(),
		baseQuantityDisplay: amount.FormatUnits(orderValue.BaseQuantity, baseDecimals),
		priceRaw:            orderValue.Price.String(),
		priceDisplay:        amount.FormatUnits(orderValue.Price, quoteDecimals),
	}, nil
}

func outputGet(cmd *cobra.Command, out *getOut) error {
	if !out.active {
		fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", out.status)
		return nil
	}

	sideDetail := ""
	if out.side == "sell" && out.quoteSymbol != "" && out.baseSymbol != "" {
		sideDetail = fmt.Sprintf(" (%s -> %s)", out.quoteSymbol, out.baseSymbol)
	} else if out.side == "buy" && out.baseSymbol != "" && out.quoteSymbol != "" {
		sideDetail = fmt.Sprintf(" (%s -> %s)", out.baseSymbol, out.quoteSymbol)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", out.status)
	fmt.Fprintf(cmd.OutOrStdout(), "Maker: %s\n", out.user)
	fmt.Fprintf(cmd.OutOrStdout(), "Base Token: %s\n", out.baseToken)
	fmt.Fprintf(cmd.OutOrStdout(), "Quote Token: %s\n", out.quoteToken)
	fmt.Fprintf(cmd.OutOrStdout(), "Side: %s%s\n", out.side, sideDetail)
	fmt.Fprintf(cmd.OutOrStdout(), "Base Quantity: %s (raw: %s)\n", out.baseQuantityDisplay, out.baseQuantityRaw)
	fmt.Fprintf(cmd.OutOrStdout(), "Price: %s (raw: %s)\n", out.priceDisplay, out.priceRaw)
	return nil
}
