package cmd

import (
	"fmt"
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

func NewRootCmd() (*cobra.Command, error) {
	formattedVersion := version.Format(build.Version, build.Date)
	rootCmd := &cobra.Command{
		Use:          "pomo <command> <subcommand> [flags]",
		Short:        "pomodoro cli",
		Long:         "simple todo list with pomodoro timer tool",
		Version:      formattedVersion,
		SilenceUsage: true,
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
