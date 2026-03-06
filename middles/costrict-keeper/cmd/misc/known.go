package client

import (
	"fmt"
	"os"
	"path/filepath"

	"costrict-keeper/cmd/root"
	"costrict-keeper/internal/env"

	"github.com/spf13/cobra"
)

var knownCmd = &cobra.Command{
	Use:   "known",
	Short: "View well-known.json file",
	Long:  "View $HOME/.costrict/share/.well-known.json",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		showKnowledge()
	},
}

func showKnowledge() {
	fname := filepath.Join(env.CostrictDir, "share", ".well-known.json")

	bytes, err := os.ReadFile(fname)
	if err != nil {
		fmt.Printf("Load '%s' failed: %v", fname, err)
		return
	}
	fmt.Printf("%s\n", string(bytes))
}

func init() {
	root.RootCmd.AddCommand(knownCmd)
}
