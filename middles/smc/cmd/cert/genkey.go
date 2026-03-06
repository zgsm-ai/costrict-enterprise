/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package cert

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/internal/utils"
)

func genKeys() {
	if err := utils.GenKeyFiles(optPublicKey, optPrivateKey); err != nil {
		fmt.Println(err)
	}
}

// genkeyCmd represents the 'smc genkey' command
var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "Generate a pair of public/private keys",
	Long:  `Generate a pair of public/private keys and saves them as ASN.1 DER encoded files`,

	Run: func(cmd *cobra.Command, args []string) {
		genKeys()
	},
}

var optPublicKey string
var optPrivateKey string

func init() {
	certCmd.AddCommand(genkeyCmd)

	genkeyCmd.Example = `  # Generate a pair of public/private key files, output public key as public.key and private key as private.pem
  smc cert genkey`
	genkeyCmd.Flags().SortFlags = false
	genkeyCmd.Flags().StringVarP(&optPublicKey, "public", "c", "public.key", "public key file")
	genkeyCmd.Flags().StringVarP(&optPrivateKey, "private", "e", "private.pem", "private key file")
}
