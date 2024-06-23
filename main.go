package main

import (
	"github.com/rcleveng/gcp_iam_search/cmd"
	_ "github.com/rcleveng/gcp_iam_search/cmd/createdb"
	_ "github.com/rcleveng/gcp_iam_search/cmd/server"
)

func main() {
	cmd.Execute()
}
