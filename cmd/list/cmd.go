package list

import (
	"github.com/aelnahas/pomo/output"
	"github.com/aelnahas/pomo/task"
	"github.com/spf13/cobra"
)

type options struct {
	all     bool
	current bool
}

func NewCmd(version string, store task.Store) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list tasks",
		Long:    "list tasks",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			var tasks []task.Task
			var err error
			if opts.all {
				tasks, err = store.List(func(t task.Task) bool {
					return true
				})
			} else if opts.current {
				task, err := store.GetCurrentTask()
				if err != nil {
					return err
				}
				tasks = append(tasks, *task)
			} else {
				tasks, err = store.List(func(t task.Task) bool {
					return t.Status != task.Complete
				})
			}

			if err != nil {
				return err
			}
			output.Printlist(tasks...)
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&opts.all, "all", "a", false, "list all tasks regardless of their status")
	cmd.PersistentFlags().BoolVarP(&opts.current, "current", "c", false, "fetch the tasks set to current")
	return cmd
}
