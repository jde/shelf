package cmdquery

import (
	"log"

	"github.com/spf13/cobra"
)

var executeLong = `Executes a query from the system by the query name.

Example:
	query execute -n "user_advice"

	query execute -n "my_query" -v "key:value,key:value"
`

// execute contains the state for this command.
var execute struct {
	name   string
	params string
}

// addExecute handles the execution of queries.
func addExecute() {
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Executes a query by name.",
		Long:  executeLong,
		Run:   runExecute,
	}

	cmd.Flags().StringVarP(&execute.name, "name", "n", "", "Name of query.")
	cmd.Flags().StringVarP(&execute.params, "vars", "v", "", "Variables required by query.")

	queryCmd.AddCommand(cmd)
}

// runExecute is the code that implements the execute command.
func runExecute(cmd *cobra.Command, args []string) {
	log.Fatal("commands", "runExecute", "The query execute functionality is still pending")
}
