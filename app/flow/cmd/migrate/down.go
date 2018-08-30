package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DownCmd generates sql migration files
var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Apply one or more of the 'down' migrations.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("execute down migration")
		return nil
	},
}