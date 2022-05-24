package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
)

type options struct {
	list bool
}

func NewCmd(version string, c *Config) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:     "config <command> [flags]",
		Short:   "set and show app configuration",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.list {
				v := reflect.ValueOf(*c)
				for i := 0; i < v.NumField(); i++ {
					fmt.Printf("%s = %v\n", v.Type().Field(i).Name, v.Field(i).Interface())
				}
			} else {
				key := args[0]
				value := args[1]
				return Update(key, value, c)
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&opts.list, "list", "l", false, "list the configs")
	return cmd
}
