package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	root "github.com/rcleveng/gcp_iam_map/cmd"

	"github.com/spf13/cobra"
)

var (
	port int16
)

func init() {
	root.RootCmd.AddCommand(serverCmd)
	serverCmd.Flags().Int16VarP(&port, "port", "p", 3000, "port to listen on")
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts API and HTTP Server",
	Long:  `Starts API and HTTP Server.`,
	RunE:  serverCommand,
}

type RolePermissions struct {
	Role       string `json:"role,omitempty"`
	Permission string `json:"permission,omitempty"`
}

func query(db *sql.DB, sql string, bindVar string) ([]RolePermissions, error) {
	ctx := context.Background()
	rps := make([]RolePermissions, 0, 100)
	rows, err := db.QueryContext(ctx, sql, bindVar)
	if err != nil {
		return nil, fmt.Errorf("error inserting role: %w", err)
	}
	defer rows.Close()

	fmt.Printf("Rows: %#v\n", rows)

	for rows.Next() {
		var rp RolePermissions
		err = rows.Scan(&rp.Role, &rp.Permission)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		fmt.Printf("%s: %s\n", rp.Role, rp.Permission)
		rps = append(rps, rp)
	}

	return rps, nil
}

func queryRoles(db *sql.DB, part string) ([]RolePermissions, error) {

	return query(db, `
		SELECT r.name, p.permission
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE r.name LIKE ?`,
		"roles/%"+part+"%")
}

func queryPermissions(db *sql.DB, part string) ([]RolePermissions, error) {

	return query(db, `
		SELECT r.name, p.permission
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE p.permission LIKE ?`,
		"%"+part+"%")
}

func queryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	w.Header().Set("Content-Type", "application/json")

	var perm []RolePermissions
	var err error
	qr := r.URL.Query().Get("qr")
	qp := r.URL.Query().Get("qp")

	if len(qr) > 0 {
		perm, err = queryRoles(db, qr)
	} else if len(qp) > 0 {
		perm, err = queryPermissions(db, qp)
	} else {
		err = fmt.Errorf("no query parameter 'qr' or 'qp' provided")
	}

	if err != nil {
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	b, err := json.MarshalIndent(perm, "", "  ")
	if err != nil {
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	fmt.Fprint(w, string(b))
}

func serverCommand(cmd *cobra.Command, args []string) error {

	fmt.Printf("server called\n\n")

	db, err := sql.Open("sqlite3", root.DbName)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		queryHandler(w, r, db)
	})
	http.Handle("/", http.FileServer(http.Dir("html/")))

	port := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(port, nil))

	return nil
}
