package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func runBase(cmd *cobra.Command, args []string) {
	fmt.Println("Root")
}
