package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			f.Value.Set(val)
		}
	}

	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}

	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

func findCommand(root *cobra.Command, path ...string) *cobra.Command {
	cmd, _, err := root.Traverse(path)
	if err != nil {
		panic(err)
	}

	return cmd
}
