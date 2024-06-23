package createdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	iam "google.golang.org/api/iam/v1"

	// Add this import
	_ "github.com/mattn/go-sqlite3"
	"github.com/rcleveng/gcp_iam_search/cmd"
	"github.com/spf13/cobra"
)

var (
	overwrite bool
)

func init() {
	cmd.RootCmd.AddCommand(CreateDbCommand)
	CreateDbCommand.Flags().BoolVarP(&overwrite, "overrite", "o", false, "Overwrite existing database")
}

// CreateDbCommand represents the createdb command
var CreateDbCommand = &cobra.Command{
	Use:   "createdb",
	Short: "Creates a sqlite database from IAM roles and permissions",
	Long:  `Creates a sqlite database from IAM roles and permissions`,
	RunE:  createDbCmdRun,
}

func handleRolePage(ctx context.Context, db *sql.DB, response *iam.ListRolesResponse) error {
	for _, role := range response.Roles {
		result, err := db.ExecContext(ctx, "INSERT INTO roles (name, title, description) VALUES (?, ?, ?)", role.Name, role.Title, role.Description)
		if err != nil {
			return fmt.Errorf("error inserting role: %w", err)
		}
		roleId, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("error getting last role id: %w", err)
		}

		for _, permission := range role.IncludedPermissions {
			result, err := db.ExecContext(ctx, "INSERT INTO permissions(permission) VALUES (?)", permission)
			if err != nil {
				return fmt.Errorf("error inserting permission: %w", err)
			}
			permissionId, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("error getting last permission id: %w", err)
			}

			_, err = db.ExecContext(ctx, "INSERT INTO role_permissions(role_id, permission_id) VALUES (?, ?)", roleId, permissionId)
			if err != nil {
				return fmt.Errorf("error inserting role_permission: %w", err)
			}
		}

		// fmt.Printf("Role: %#v\nPermissions: %#v\n", role.Name, role.IncludedPermissions)
		fmt.Print(".")
	}
	return nil
}

func createTable(ctx context.Context, db *sql.DB, tableName string, schema string) error {
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("error: [%w] creating table: [%s] ", err, tableName)
	}
	return nil
}

func createTables(ctx context.Context, db *sql.DB) error {
	if err := createTable(ctx, db, "roles", `
CREATE TABLE [roles] ( [id] INTEGER PRIMARY KEY AUTOINCREMENT, [name] TEXT, [title] TEXT, [description] TEXT)
	`); err != nil {
		return fmt.Errorf("error creating roles table: %w", err)
	}
	fmt.Println("Roles table created successfully")

	if err := createTable(ctx, db, "permissions", `
CREATE TABLE [permissions] ( [id] INTEGER PRIMARY KEY AUTOINCREMENT, [permission] TEXT)
	`); err != nil {
		return fmt.Errorf("error creating permissions table: %w", err)
	}
	fmt.Println("permissions table created successfully")

	if err := createTable(ctx, db, "roles_permissions", `
CREATE TABLE [role_permissions] (
  [role_id] INTEGER NOT NULL,
  [permission_id] INTEGER NOT NULL,
  PRIMARY KEY (role_id, permission_id),
  FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
	`); err != nil {
		return fmt.Errorf("error creating roles table: %w", err)
	}
	fmt.Println("Roles table created successfully")
	return nil
}

func createDbCmdRun(cmd *cobra.Command, args []string) error {
	fmt.Println("createdb called")
	if len(args) == 0 {
		return fmt.Errorf("please provide a filename")
	}
	filename := args[0]
	// generate: if file exists
	if _, err := os.Stat(filename); err == nil {
		if !overwrite {
			return fmt.Errorf("file %s already exists", filename)
		}
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("error removing file: %w", err)
		}
	}

	// Example code to read IAM roles and permissions from GCP
	ctx := cmd.Context()
	c, err := iam.NewService(ctx)
	if err != nil {
		return fmt.Errorf("error creating IAM client: %w", err)
	}

	// Create database
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Create all expected tables
	if err := createTables(ctx, db); err != nil {
		return fmt.Errorf("error creating tables: %w", err)
	}

	call := c.Roles.List().Context(ctx).PageSize(1000).View("FULL")
	if err := call.Pages(ctx, func(lrr *iam.ListRolesResponse) error {
		return handleRolePage(ctx, db, lrr)
	}); err != nil {
		return fmt.Errorf("error listing roles: %v", err)
	}

	fmt.Println("SQLite database created successfully")
	return nil
}
