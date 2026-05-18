// Copyright © 2026 0xTrooper (on Github)
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	configcmd "dex/cmd/config"
	tradecmd "dex/cmd/trade"
	tokencmd "dex/cmd/token"
	walletcmd "dex/cmd/wallet"
	testcmd "dex/cmd/test"
	"dex/service"

	"github.com/spf13/cobra"
)

func NewCommand(svc *service.Service, ks *service.Keystore) *cobra.Command {
	root := &cobra.Command{
		Use:   "dex",
		Short: "Decentralized Exchange CLI",
		Long:  `Decentralized Exchange Command Line Interface.`,
	}

	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(configcmd.NewCommand(svc))
	root.AddCommand(walletcmd.NewCommand(svc, ks))
	root.AddCommand(tradecmd.NewCommand(svc, ks))
	root.AddCommand(tokencmd.NewCommand(svc, ks))
	root.AddCommand(testcmd.NewCommand(svc, ks))

	return root
}

func Execute() {
	svc, err := service.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}

	ks, err := service.NewKeystore()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}

	if err := NewCommand(svc, ks).Execute(); err != nil {
		os.Exit(exitFailure)
	}
}
