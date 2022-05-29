package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aelnahas/pomo/build"
	"github.com/aelnahas/pomo/cmd/add"
	"github.com/aelnahas/pomo/cmd/config"
	"github.com/aelnahas/pomo/cmd/list"
	"github.com/aelnahas/pomo/cmd/remove"
	"github.com/aelnahas/pomo/cmd/set"
	"github.com/aelnahas/pomo/cmd/timer"
	"github.com/aelnahas/pomo/cmd/version"
	"github.com/aelnahas/pomo/sessions"
	"github.com/aelnahas/pomo/task"
	"github.com/spf13/cobra"
)

type options struct {
	init bool
}

func NewRootCmd() (*cobra.Command, error) {
	opts := options{}
	formattedVersion := version.Format(build.Version, build.Date)
	rootCmd := &cobra.Command{
		Use:          "pomo <command> <subcommand> [flags]",
		Short:        "pomodoro cli",
		Long:         "simple todo list with pomodoro timer tool",
		Version:      formattedVersion,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.init {
				src := config.Template
				dir, err := config.ExpandPath(config.PomoDir)
				if err != nil {
					return err
				}
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					return err
				}

				fin, err := os.Open(src)
				if err != nil {
					return err
				}
				defer fin.Close()

				tgt := fmt.Sprintf("%s/%s", dir, config.PomoConfig)
				if _, err := os.Stat(tgt); errors.Is(err, os.ErrNotExist) {
					fout, err := os.Create(tgt)
					if err != nil {
						return err
					}
					defer fout.Close()
					_, err = io.Copy(fout, fin)
					return err
				}

				return fmt.Errorf("config in: %s already exists", tgt)
			}

			return nil
		},
	}

	appConfig, err := config.Parse(config.DefaultPath)
	if err != nil {
		return nil, err
	}

	taskPath, err := config.ExpandPath(appConfig.Database.Task)
	if err != nil {
		return nil, err
	}

	store, err := task.NewStore(taskPath)
	if err != nil {
		return nil, err
	}

	sessionPath, err := config.ExpandPath(appConfig.Database.Session)
	if err != nil {
		return nil, err
	}

	sessionStore, err := sessions.NewStore(sessionPath, appConfig.Timers.Interval)
	if err != nil {
		return nil, err
	}

	rootCmd.SetVersionTemplate(formattedVersion)
	rootCmd.AddCommand(version.NewCmd(build.Version, build.Date))
	rootCmd.AddCommand(add.NewCmd(formattedVersion, store))
	rootCmd.AddCommand(set.NewCmd(formattedVersion, store))
	rootCmd.AddCommand(timer.NewCmd(formattedVersion, appConfig, store, sessionStore))
	rootCmd.AddCommand(config.NewCmd(formattedVersion, appConfig))
	rootCmd.AddCommand(list.NewCmd(formattedVersion, store))
	rootCmd.AddCommand(remove.NewCmd(formattedVersion, store))
	rootCmd.PersistentFlags().BoolVar(&opts.init, "init", false, "initialize default config")
	return rootCmd, nil
}

func Execute() {
	rootCmd, err := NewRootCmd()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(2)
	}

	if rootCmd.Execute(); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
