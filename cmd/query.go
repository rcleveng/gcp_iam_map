package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "queries a sqlite database for permissions/roles",
	Long: `queries a sqlite database for permissions/roles. For example:

gcp_iam_map queries an existing database for roles and permissions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("query called")
	},
}

func init() {
	RootCmd.AddCommand(queryCmd)

	// queryCmd.PersistentFlags().String("query", "", "A help for query")

	// queryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
