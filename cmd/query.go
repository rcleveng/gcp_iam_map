package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/rcleveng/gcp_iam_search/pkg/iamdb"
	"github.com/spf13/cobra"
)

var (
	cpuProfile string
	wildcard   bool
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "queries a sqlite database for permissions/roles",
	Long: `queries a sqlite database for permissions/roles. For example:

gcp_iam_search queries an existing database for roles and permissions.`,
	RunE: queryCommand,
}

func queryCommand(cmd *cobra.Command, args []string) error {

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %w", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if len(args) == 0 {
		return fmt.Errorf("query string is empty")
	}
	queryString := args[0]

	dbName, err := cmd.Flags().GetString("db")
	if err != nil {
		return fmt.Errorf("error getting database name: %w", err)
	}

	db, err := iamdb.NewIamDB(dbName)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	results, err := db.QueryPermissions(queryString, wildcard)
	if err != nil {
		return fmt.Errorf("error executing database query: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	if wildcard {
		for _, rp := range results {
			fmt.Printf("%s: %s\n", rp.Role, rp.Permission)
		}
	} else {
		for _, rp := range results {
			fmt.Printf("%s\n", rp.Role)
		}

	}

	return nil
}

func init() {
	RootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringVarP(&cpuProfile, "cpu", "c", "", "cpu profile")
	queryCmd.Flags().BoolVar(&wildcard, "wildcard", true, "use wildcard search")
}
