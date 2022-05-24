package timer

import (
	"fmt"
	"time"

	"github.com/aelnahas/pomo/cmd/config"
	"github.com/aelnahas/pomo/countdown"
	"github.com/aelnahas/pomo/output"
	"github.com/aelnahas/pomo/sessions"
	"github.com/aelnahas/pomo/task"
	"github.com/spf13/cobra"
)

type options struct {
	reset bool
	show  bool
}

func NewCmd(version string, config *config.Config, store task.Store, sessionStore sessions.Store) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:     "timer <command> [flags]",
		Aliases: []string{"t"},
		Short:   "control pomo timer",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.reset {
				return sessionStore.Reset()
			}

			if opts.show {
				session, err := sessionStore.Session()
				if err != nil {
					return err
				}

				output.PrintSession(session.Current, session.Next(config.Timers.Interval), session.Count)
			}

			return nil
		},
	}

	cmd.AddCommand(newStartCmd(version, config, store, sessionStore))
	cmd.PersistentFlags().BoolVarP(&opts.reset, "reset", "r", false, "reset sessions")
	cmd.PersistentFlags().BoolVarP(&opts.show, "show", "s", false, "show sessions")
	return cmd
}

func newStartCmd(version string, config *config.Config, store task.Store, sessionStore sessions.Store) *cobra.Command {
	return &cobra.Command{
		Use:     "start",
		Short:   "start a timer",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			current, err := store.GetCurrentTask()
			if err != nil {
				return fmt.Errorf("current task is not set (%w)", err)
			}

			var duration time.Duration
			sessionType, err := sessionStore.Current()
			if err != nil {
				return err
			}

			switch sessionType {
			case sessions.Focus:
				duration = config.Timers.FocusDuration()
			case sessions.Short:
				duration = config.Timers.ShortBreakDuration()
			default:
				duration = config.Timers.LongBreakDuration()
			}

			if err := countdown.New(duration, current, sessionType).Run(); err != nil {
				return err
			}

			if sessionType == sessions.Focus {
				current, err = store.AddSessions(current.ID)
				if err != nil {
					return err
				}
			}

			if err := sessionStore.Increment(); err != nil {
				return err
			}

			output.Printlist(*current)
			return nil
		},
	}
}
