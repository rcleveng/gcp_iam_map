package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
)

var (
	cpuProfile string
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "queries a sqlite database for permissions/roles",
	Long: `queries a sqlite database for permissions/roles. For example:

gcp_iam_search queries an existing database for roles and permissions.`,
	RunE: queryCommand,
}

func query(db *sql.DB, queryString string) error {
	fmt.Printf("Searching for IAM permision substring: '%s'\n\n", queryString)

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `
SELECT r.name, p.permission
   FROM roles r
   JOIN role_permissions rp ON r.id = rp.role_id
   JOIN permissions p ON rp.permission_id = p.id
   WHERE p.permission LIKE ?`, "%"+queryString+"%")
	if err != nil {
		return fmt.Errorf("error inserting role: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var roleName, permission string
		err = rows.Scan(&roleName, &permission)
		if err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}
		fmt.Printf("%s: %s\n", roleName, permission)
		count++
	}

	if count == 0 {
		fmt.Println("No results found.")
	}

	return nil
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

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	err = query(db, queryString)
	if err != nil {
		return fmt.Errorf("error querying database: %w", err)
	}

	return nil
}

func init() {
	RootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringVarP(&cpuProfile, "cpu", "c", "", "cpu profile")
}
