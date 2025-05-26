/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package key

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"fmt"
)

// setCmd represents the set command
var availableProviders = []string{"gemini", "luminai"}

var setCmd = &cobra.Command{
	Use:   "set [provider]",
	Short: "Set AI provider",
	Long:  `Set the preferred AI provider (e.g. gemini, openai, luminai)`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]

		// Validasi apakah provider termasuk yang tersedia
		if !isValidProvider(provider) {
			fmt.Printf("❌ Invalid provider: '%s'. Available options: %v\n", provider, availableProviders)
			return
		}

		viper.Set("ai.provider", provider)
		cobra.CheckErr(viper.WriteConfig())

		fmt.Printf("✅ AI provider set to: %s\n", provider)
	},
}

func isValidProvider(provider string) bool {
	for _, p := range availableProviders {
		if p == provider {
			return true
		}
	}
	return false
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
