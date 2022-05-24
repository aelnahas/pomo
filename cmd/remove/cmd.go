package remove

import (
	"github.com/aelnahas/pomo/task"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewCmd(version string, store task.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <task-id>",
		Short:   "remove a task",
		Long:    "remove a task",
		Aliases: []string{"rm"},
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			idString := args[0]
			id, err := uuid.Parse(idString)
			if err != nil {
				return err
			}
			return store.Remove(id)
		},
	}

	return cmd
}
