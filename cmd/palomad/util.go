package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) error {
	set := func(s *pflag.FlagSet, key, val string) error {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			if err := f.Value.Set(val); err != nil {
				return err
			}
		}
		return nil
	}

	for key, val := range defaults {
		if err := set(c.Flags(), key, val); err != nil {
			return err
		}
		if err := set(c.PersistentFlags(), key, val); err != nil {
			return err
		}
	}

	for _, c := range c.Commands() {
		return overwriteFlagDefaults(c, defaults)
	}
	return nil
}

func findCommand(root *cobra.Command, path ...string) *cobra.Command {
	cmd, _, err := root.Traverse(path)
	if err != nil {
		panic(err)
	}

	return cmd
}
