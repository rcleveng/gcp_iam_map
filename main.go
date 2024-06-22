package main

import (
	"github.com/rcleveng/gcp_iam_map/cmd"
	_ "github.com/rcleveng/gcp_iam_map/cmd/maker"
)

func main() {
	cmd.Execute()
}
