package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	root "github.com/rcleveng/gcp_iam_search/cmd"
	"github.com/rcleveng/gcp_iam_search/pkg/iamdb"

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

func queryHandler(w http.ResponseWriter, r *http.Request, db *iamdb.IamDB) {
	w.Header().Set("Content-Type", "application/json")

	var perm []iamdb.RolePermissions
	var err error
	qr := r.URL.Query().Get("qr")
	qp := r.URL.Query().Get("qp")
	wildcard, err := strconv.ParseBool(r.URL.Query().Get("wildcard"))
	if err != nil {
		wildcard = false
	}

	if len(qr) > 0 {
		perm, err = db.QueryRoles(qr, wildcard)
	} else if len(qp) > 0 {
		perm, err = db.QueryPermissions(qp, wildcard)
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
	db, err := iamdb.NewIamDB(root.DbName)
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}
	defer db.Close()

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		queryHandler(w, r, db)
	})
	http.Handle("/", http.FileServer(http.Dir("html/")))

	port := fmt.Sprintf(":%d", port)
	fmt.Printf("Server running at URL: http://localhost%s/\n\n", port)
	log.Fatal(http.ListenAndServe(port, nil))

	return nil
}
