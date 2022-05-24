package version

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func NewCmd(version, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "show version",
		Long:  "show version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Format(version, date))
		},
	}
}

func Format(version, date string) string {
	version = strings.TrimPrefix(version, "v")
	if date != "" {
		date = fmt.Sprintf(" (%s)", date)
	}

	return fmt.Sprintf("pomo version %s%s\n", version, date)
}
