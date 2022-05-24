package set

import (
	"github.com/aelnahas/pomo/output"
	"github.com/aelnahas/pomo/task"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type options struct {
	complete bool
	current  bool
}

func NewCmd(version string, store task.Store) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:     "set <title-id> [flags]",
		Short:   "set the status of a task",
		Long:    "set a task to complete, or as current one to work on",
		Aliases: []string{"s", "set-status"},
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}

			if opts.current {
				if err := store.SetCurrentTask(id); err != nil {
					return err
				}
			}

			if opts.complete {
				updated, err := store.SetState(id, task.Complete)
				if err != nil {
					return err
				}
				if err := store.ClearCurrentTask(id); err != nil {
					return err
				}

				output.Printlist(*updated)
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&opts.complete, "complete", false, "set the task complete")
	cmd.PersistentFlags().BoolVar(&opts.current, "current", false, "set current task to use with timer")
	return cmd
}
