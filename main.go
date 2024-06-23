package main

import (
	"github.com/rcleveng/gcp_iam_map/cmd"
	_ "github.com/rcleveng/gcp_iam_map/cmd/maker"
	_ "github.com/rcleveng/gcp_iam_map/cmd/server"
)

func main() {
	cmd.Execute()
}
