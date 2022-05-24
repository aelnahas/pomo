package add

import (
	"github.com/aelnahas/pomo/output"
	"github.com/aelnahas/pomo/task"
	"github.com/spf13/cobra"
)

func NewCmd(version string, store task.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <title>",
		Aliases: []string{"a"},
		Short:   "add a new task",
		Example: "add 'title for my new task'",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			newTask, err := store.Add(args[0])
			if err != nil {
				return err
			}
			output.Printlist(*newTask)
			return nil
		},
	}

	return cmd
}
