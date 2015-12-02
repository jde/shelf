package cmdquery

import (
	"os"
	"path/filepath"

	"github.com/coralproject/shelf/pkg/cfg"
	"github.com/coralproject/shelf/pkg/db"
	"github.com/coralproject/shelf/pkg/log"
	"github.com/coralproject/shelf/pkg/srv/query"

	"github.com/spf13/cobra"
)

var updateLong = `Use update to change a query in the system.
Updating can be done per file or per directory.

Note: Update will check for a $SHELF_PATH environment variable of which it
appends a './queries' to, when no dirPath or fileName is given.

Example:
	To update a single file:
	query update -p user_advice.json

	To update using a directory:
	query update -p ./queries

	To load using the environment variable path:
	query update
`

// update contains the state for this command.
var update struct {
	path string
}

// addUpd handles the update of query record.
func addUpd() {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates a query from a file or directory.",
		Long:  updateLong,
		Run:   runUpdate,
	}

	cmd.Flags().StringVarP(&update.path, "path", "p", "", "Path of file or directory.")

	queryCmd.AddCommand(cmd)
}

// runUpdate is the code that implements the create command.
func runUpdate(cmd *cobra.Command, args []string) {
	if update.path == "" {
		dir, err := cfg.String(envKey)
		if err != nil {
			update.path = defDir
		} else {
			update.path = filepath.Join(dir, defDir)
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Error("commands", "runUpdate", err, "Completed")
		return
	}

	file := filepath.Join(pwd, update.path)

	stat, err := os.Stat(file)
	if err != nil {
		log.Error("commands", "runUpdate", err, "Completed")
		return
	}

	db := db.NewMGO()
	defer db.CloseMGO()

	if !stat.IsDir() {
		q, err := setFromFile("commands", file)
		if err != nil {
			log.Error("commands", "runUpdate", err, "Completed")
			return
		}

		if err := query.CreateSet("commands", db, q); err != nil {
			log.Error("commands", "runUpdate", err, "Completed")
			return
		}

		return
	}

	err2 := loadDir(file, func(path string) error {
		q, err := setFromFile("commands", path)
		if err != nil {
			return err
		}

		return query.CreateSet("commands", db, q)
	})

	if err2 != nil {
		log.Error("commands", "runUpdate", err, "Completed")
		return
	}
}
